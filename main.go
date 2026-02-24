package main

import (
	"log"
	mon "nginx-log-monitor" // Import the local module
	"nginx-log-monitor/configPkg"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := configPkg.LoadConfig("nginx-log-monitor/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize the monitor with log file path and config flag
	// Assuming access.log exists or will be created
	monitor, err := mon.NewMonitor("./access.log", cfg.Database.ClearOnStartup)
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}
	monitor.Start()

	router := gin.Default()
	v1 := router.Group("/api/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		// Expose stats endpoint
		v1.GET("/stats", func(c *gin.Context) {
			c.JSON(200, monitor.GetStats())
		})
	}

	router.Run() // listens on 0.0.0.0:8080 by default
}
