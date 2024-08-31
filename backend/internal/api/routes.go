package api

import (
	"ats-project/backend/internal/scpi"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, scpiClient *scpi.Client) {
	r.GET("/ws", func(c *gin.Context) {
		HandleWebSocket(c.Writer, c.Request, scpiClient)
	})
}
