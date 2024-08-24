package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

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

type StartMessage struct {
	Action   string `json:"action"`
	Devices  []int  `json:"devices"`
	Channels []int  `json:"channels"`
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
	//log.Printf("Received message: %s\n", message)

	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("Error parsing JSON message:", err)
		return
	}

	action, ok := msg["action"].(string)
	if !ok {
		log.Println("Invalid action in message")
		return
	}

	switch action {
	case "start":
		var startMsg StartMessage
		if err := json.Unmarshal(message, &startMsg); err != nil {
			log.Println("Error parsing start message:", err)
			return
		}
		startMeasurement(conn, scpiClient, startMsg.Devices, startMsg.Channels)
	case "stop":
		stopMeasurement(scpiClient)
	default:
		log.Println("Unknown action:", action)
	}
}

func startMeasurement(conn *websocket.Conn, scpiClient *scpi.Client, devices, channels []int) {
	log.Println("Measurement started")
	for _, device := range devices {
		for _, channel := range channels {
			//log.Printf("Measuring voltage for device %d, channel %d\n", device, channel)
			voltages, err := scpiClient.MeasureVoltage(device, channel)
			if err != nil {
				log.Printf("Error measuring voltage for device %d, channel %d: %v\n", device, channel, err)
				continue
			}

			passed := true
			for _, voltage := range voltages {
				if voltage >= 6554 && voltage <= 45875 { //0.5v-3.5v
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

			jsonResult, err := json.Marshal(result)
			if err != nil {
				log.Printf("Error marshaling result: %v\n", err)
				continue
			}

			//log.Printf("JSON result: %s\n", jsonResult)
			if err := conn.WriteMessage(websocket.TextMessage, jsonResult); err != nil {
				log.Printf("Error sending result: %v\n", err)
			}

			// 插入延时
			//time.Sleep(200 * time.Millisecond) // 这里设置延时为1秒，可以根据需要调整
		}
	}
}

func stopMeasurement(scpiClient *scpi.Client) {
	log.Println("Measurement stopped")
	if err := scpiClient.StopMeasurement(); err != nil {
		log.Printf("Error stopping measurement: %v\n", err)
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
