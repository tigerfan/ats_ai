package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"ats-project/backend/internal/api"
	"ats-project/backend/internal/db"
	"ats-project/backend/internal/scpi"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Measurement struct {
		Devices int `json:"devices"`
	} `json:"measurement"`
	ScpiServer struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"scpi_server"`
	HttpServer struct {
		Port int `json:"port"`
	} `json:"http_server"`
}

func main() {
	// Read configuration
	config, err := readConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to read config: %v", err)
	}

	// Initialize database connection
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize SCPI client
	scpiClient := scpi.NewClient()
	err = scpiClient.Connect(config.ScpiServer.Host, config.ScpiServer.Port, config.Measurement.Devices)
	if err != nil {
		log.Fatalf("Failed to connect to SCPI servers: %v", err)
	}
	defer scpiClient.Close()

	// Create Gin engine
	r := gin.Default()

	// Set up routes
	api.SetupRoutes(r, scpiClient)

	// Start HTTP server (including WebSocket)
	if err := r.Run(fmt.Sprintf(":%d", config.HttpServer.Port)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func readConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open config file %s: %v", filename, err)
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("could not decode config file %s: %v", filename, err)
	}

	return &config, nil
}
