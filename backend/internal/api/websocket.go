package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"ats-project/backend/internal/db"
	"ats-project/backend/internal/scpi"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

type Message struct {
	Action   string      `json:"action"`
	Devices  []int       `json:"devices"`
	Channels []int       `json:"channels"`
	Params   interface{} `json:"params"`
}

var (
	measurementRunning bool
	measurementPaused  bool
	pauseCond          *sync.Cond
	stateMutex         sync.Mutex
	stopChan           chan struct{}
)

func init() {
	pauseCond = sync.NewCond(&stateMutex)
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request, scpiClient *scpi.Client) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading WebSocket message:", err)
			break
		}
		handleWebSocketMessage(conn, messageType, p, scpiClient)
	}

	clientsMu.Lock()
	delete(clients, conn)
	clientsMu.Unlock()
}

func handleWebSocketMessage(conn *websocket.Conn, messageType int, message []byte, scpiClient *scpi.Client) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error parsing JSON message:", err)
		return
	}

	stateMutex.Lock()
	defer stateMutex.Unlock()

	switch msg.Action {
	case "start":
		if !measurementRunning {
			measurementRunning = true
			measurementPaused = false
			stopChan = make(chan struct{})
			go startMeasurement(conn, scpiClient, msg.Devices, msg.Channels)
		}
	case "pause":
		if measurementRunning && !measurementPaused {
			measurementPaused = true
		}
	case "resume":
		if measurementRunning && measurementPaused {
			measurementPaused = false
			pauseCond.Broadcast()
		}
	case "stop":
		if measurementRunning {
			if stopChan != nil {
				select {
				case <-stopChan:
					// Channel is already closed, do nothing
				default:
					close(stopChan)
				}
			}
			measurementRunning = false
			measurementPaused = false
			pauseCond.Broadcast()
		}
	case "getMeasurementHistory":
		go sendMeasurementHistory(conn)
	case "getHistoricalData":
		go sendHistoricalData(conn, msg.Params)
	default:
		log.Println("Unknown action:", msg.Action)
	}
}

func startMeasurement(conn *websocket.Conn, scpiClient *scpi.Client, devices, channels []int) {
	log.Println("Measurement started")
	startTime := time.Now()

	results := make(chan map[string]interface{}, len(devices)*len(channels))
	allResults := []map[string]interface{}{}
	defer func() {
		stateMutex.Lock()
		measurementRunning = false
		measurementPaused = false
		if stopChan != nil {
			select {
			case <-stopChan:
				// Channel is already closed, do nothing
			default:
				close(stopChan)
			}
			stopChan = nil
		}
		stateMutex.Unlock()
		log.Println("Measurement completed")

		endTime := time.Now()

		for result := range results {
			allResults = append(allResults, result)
		}

		// Asynchronously write measurement history and data to the database
		go writeMeasurementToDB(conn, devices, channels, startTime, endTime, allResults)
	}()

	var wg sync.WaitGroup

	for _, device := range devices {
		wg.Add(1)
		go measureDevice(scpiClient, device, channels, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	sendResults(conn, results)
}

func measureDevice(scpiClient *scpi.Client, device int, channels []int, results chan<- map[string]interface{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for _, channel := range channels {
		stateMutex.Lock()
		for measurementPaused {
			pauseCond.Wait()
		}
		if !measurementRunning {
			stateMutex.Unlock()
			return
		}
		stateMutex.Unlock()

		select {
		case <-stopChan:
			return
		default:
			voltages, err := scpiClient.MeasureVoltage(device, channel)
			if err != nil {
				log.Printf("Error measuring voltage for device %d, channel %d: %v\n", device, channel, err)
				continue
			}

			passed := true
			for _, voltage := range voltages {
				if voltage > 6554 && voltage < 45875 { // 0.5v-3.5v
					passed = false
					break
				}
			}

			result := map[string]interface{}{
				"device":   device,
				"channel":  channel,
				"voltages": voltages,
				"passed":   passed,
			}

			results <- result
		}
	}
}

func sendResults(conn *websocket.Conn, results <-chan map[string]interface{}) {
	batch := make([]map[string]interface{}, 0, 10)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case result, ok := <-results:
			if !ok {
				if len(batch) > 0 {
					sendBatch(conn, batch, true)
				}
				return
			}
			batch = append(batch, result)
			if len(batch) >= 10 {
				sendBatch(conn, batch, false)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				sendBatch(conn, batch, false)
				batch = batch[:0]
			}
		}
	}
}

func sendBatch(conn *websocket.Conn, batch []map[string]interface{}, completed bool) {
	response := map[string]interface{}{
		"status":  "in_progress",
		"results": batch,
	}

	if completed {
		response["status"] = "completed"
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling response: %v\n", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonResponse); err != nil {
		log.Printf("Error sending response: %v\n", err)
	}
}

func writeMeasurementToDB(conn *websocket.Conn, devices, channels []int, startTime, endTime time.Time, results []map[string]interface{}) {
	// 写入测量历史记录
	historyID, err := db.WriteMeasurementHistory(startTime, endTime, "completed", len(devices), len(channels))
	if err != nil {
		log.Printf("Error writing measurement history to database: %v\n", err)
		sendMessage(conn, "error", "Error writing measurement history to the database")
		return
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // 限制并发写入的数量

	for _, result := range results {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(r map[string]interface{}) {
			defer wg.Done()
			defer func() { <-semaphore }()

			device := r["device"].(int)
			channel := r["channel"].(int)
			voltages, ok := r["voltages"].([]float64)
			if !ok {
				log.Printf("Error: voltages is not of type []float64\n")
				return
			}

			if err := db.WriteMeasurementData(historyID, device, channel, voltages); err != nil {
				log.Printf("Error writing measurement data to database: %v\n", err)
			}
		}(result)
	}

	wg.Wait()

	// 通知客户端写入数据库过程已完成
	sendMessage(conn, "completed", "Measurement data has been successfully written to the database")
	log.Println("Measurement data has been successfully written to the database!")
}

func sendMessage(conn *websocket.Conn, status, message string) {
	response := map[string]interface{}{
		"status":  status,
		"message": message,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshaling message: %v\n", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, jsonResponse); err != nil {
		log.Printf("Error sending message: %v\n", err)
	}
}

func sendHistoricalData(conn *websocket.Conn, params interface{}) {
	// 解析参数以获取 historyID
	paramsMap, ok := params.(map[string]interface{})
	if !ok {
		sendErrorMessage(conn, "Invalid parameters for historical data")
		return
	}

	historyID, ok := paramsMap["historyID"].(float64)
	if !ok {
		sendErrorMessage(conn, "Invalid history ID")
		return
	}

	// 从 InfluxDB 获取历史数据
	voltages, err := db.GetHistoricalData(int64(historyID))
	if err != nil {
		sendErrorMessage(conn, "Error fetching historical data")
		return
	}

	response := map[string]interface{}{
		"status":    "historicalData",
		"historyID": int64(historyID),
		"results":   voltages,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending historical data: %v\n", err)
	}
}

func sendMeasurementHistory(conn *websocket.Conn) {
	// 从 InfluxDB 获取测量历史记录
	history, err := db.GetMeasurementHistory()
	if err != nil {
		sendErrorMessage(conn, "Error fetching measurement history")
		return
	}

	response := map[string]interface{}{
		"status":  "measurementHistory",
		"history": history,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending measurement history: %v\n", err)
	}
}

func sendErrorMessage(conn *websocket.Conn, message string) {
	response := map[string]interface{}{
		"status":  "error",
		"message": message,
	}

	if err := conn.WriteJSON(response); err != nil {
		log.Printf("Error sending error message: %v\n", err)
	}
}

func BroadcastMessage(message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("Error broadcasting message:", err)
			client.Close()
			delete(clients, client)
		}
	}
}
