package api

import (
	"SvcWatch/internal/config"
	"SvcWatch/internal/middleware"
	"SvcWatch/internal/monitor"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes and configures the Gin API router.
func SetupRouter(monitors []*monitor.Monitor, cfg *config.Config) *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		// Public routes
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Protected routes require token authentication
		private := v1.Group("")
		private.Use(middleware.TokenAuthMiddleware(cfg.Auth.PassportURL))
		{
			// Expose combined stats endpoint for all targets
			private.GET("/stats", func(c *gin.Context) {
				stats := make(map[string]interface{})
				for i, monInst := range monitors {
					tableName := cfg.Targets[i].Table
					stats[tableName] = monInst.GetStats()["total_logs"]
				}
				c.JSON(200, stats)
			})
		}
	}

	return router
}
