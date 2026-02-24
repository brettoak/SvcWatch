package main

import (
	"log"
	"nginx-log-monitor" // Import the local module
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize the monitor with log file path
	// Assuming access.log exists or will be created
	mon, err := monitor.NewMonitor("./access.log")
	if err != nil {
		log.Fatalf("Failed to create monitor: %v", err)
	}
	mon.Start()

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Expose stats endpoint
	router.GET("/stats", func(c *gin.Context) {
		c.JSON(200, mon.GetStats())
	})

	router.Run() // listens on 0.0.0.0:8080 by default
}
