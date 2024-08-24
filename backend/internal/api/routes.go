package api

import (
	"ats-project/backend/internal/scpi"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, scpiClient *scpi.Client) {
	r.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(c.Writer, c.Request, scpiClient)
	})
	r.GET("/api/measurement-history", handleMeasurementHistory)
	r.GET("/api/historical-data", handleHistoricalData)

	// 添加更多路由...
}

func handleMeasurementHistory(c *gin.Context) {
	// 实现获取测量历史的逻辑
}

func handleHistoricalData(c *gin.Context) {
	// 实现获取历史数据的逻辑
}
