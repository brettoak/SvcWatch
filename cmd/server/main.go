package main

import (
	"SvcWatch/internal/api"
	"SvcWatch/internal/config"
	mon "SvcWatch/internal/monitor" // Import the local module
	storage "SvcWatch/internal/storage"
	"log"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
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

	// Setup and start the router
	router := api.SetupRouter(monitors, cfg)
	router.Run() // listens on 0.0.0.0:8080 by default
}
