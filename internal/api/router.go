package api

import (
	"SvcWatch/internal/config"
	"SvcWatch/internal/monitor"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin API router.
func SetupRouter(monitors []*monitor.Monitor, cfg *config.Config) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Expose combined stats endpoint for all targets
		v1.GET("/stats", func(c *gin.Context) {
			stats := make(map[string]interface{})
			for i, monInst := range monitors {
				tableName := cfg.Targets[i].Table
				stats[tableName] = monInst.GetStats()["total_logs"]
			}
			c.JSON(200, stats)
		})
	}

	return router
}
