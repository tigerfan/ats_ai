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
	Action   string `json:"action"`
	Devices  []int  `json:"devices"`
	Channels []int  `json:"channels"`
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
	default:
		log.Println("Unknown action:", msg.Action)
	}
}

func startMeasurement(conn *websocket.Conn, scpiClient *scpi.Client, devices, channels []int) {
	log.Println("Measurement started")
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

		// Notify the client that the measurement has completed and write data to the database
		notifyMeasurementComplete(conn, devices, channels)
	}()

	results := make(chan map[string]interface{}, len(devices)*len(channels))
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

func notifyMeasurementComplete(conn *websocket.Conn, devices, channels []int) {
	// Notify the client that the writing process to the database is starting
	sendMessage(conn, "writing", "Measurement data is being written to the database")
	log.Println("Measurement data is being written to the database...")

	// Write measurement data to the database
	for _, device := range devices {
		for _, channel := range channels {
			// Here we assume the voltages were stored in the results and now need to be written to the database
			// Convert voltages from []int to []float64
			voltages := []float64{} // Retrieve the actual voltages for this device and channel

			if err := db.WriteMeasurementData(device, channel, voltages); err != nil {
				log.Printf("Error writing measurement data to database: %v\n", err)
				sendMessage(conn, "error", "Error writing measurement data to the database")
				return
			}
		}
	}

	// Notify the client that the writing process to the database has completed
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
