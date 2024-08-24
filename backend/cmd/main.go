package main

import (
	"log"

	"ats-project/backend/internal/api"
	"ats-project/backend/internal/db"
	"ats-project/backend/internal/scpi"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化数据库连接
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// 初始化SCPI客户端
	scpiClient := scpi.NewClient()
	err := scpiClient.Connect("localhost:5025") // 假设SCPI服务器在本地5025端口
	if err != nil {
		log.Fatalf("Failed to connect to SCPI server: %v", err)
	}
	defer scpiClient.Close()

	// 创建Gin引擎
	r := gin.Default()

	// 设置路由
	api.SetupRoutes(r, scpiClient)

	// 启动HTTP服务器（包括WebSocket）
	if err := r.Run(":5177"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
