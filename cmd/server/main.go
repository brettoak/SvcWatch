package main

import (
	"SvcWatch/internal/configPkg"
	mon "SvcWatch/internal/monitor" // Import the local module
	storage "SvcWatch/internal/storage"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := configPkg.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize shared storage
	store := storage.NewSqliteStorage("nginx_logs.db")
	defer store.Close()

	var monitors []*mon.Monitor

	for _, target := range cfg.Targets {
		// Initialize the table for the target
		store.InitTable(target.Table, cfg.Database.ClearOnStartup)

		// Create a monitor for each target
		monitor, err := mon.NewMonitor(target.Path, store, target.Table)
		if err != nil {
			log.Fatalf("Failed to create monitor for %s: %v", target.Path, err)
		}
		monitor.Start()
		monitors = append(monitors, monitor)
	}

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
			for i, monitor := range monitors {
				tableName := cfg.Targets[i].Table
				stats[tableName] = monitor.GetStats()["total_logs"]
			}
			c.JSON(200, stats)
		})
	}

	router.Run() // listens on 0.0.0.0:8080 by default
}
