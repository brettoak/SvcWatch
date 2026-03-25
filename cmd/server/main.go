package main

import (
	"SvcWatch/internal/api"
	"SvcWatch/internal/config"
	"SvcWatch/internal/controller"
	mon "SvcWatch/internal/monitor" // Import the local module
	"SvcWatch/internal/service"
	storage "SvcWatch/internal/storage"
	"fmt"
	"log"
	"os"

	_ "SvcWatch/docs"
)

// @title SvcWatch API
// @version 1.0
// @description SvcWatch is a real-time Nginx log monitoring system.
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	// Resolve config file based on APP_ENV (development / staging / production)
	// Defaults to "development" if not set.
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	configPath := fmt.Sprintf("config/config.%s.yaml", env)
	log.Printf("Loading config: %s (APP_ENV=%s)", configPath, env)
	cfg, err := config.LoadConfig(configPath)
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

	// Initialize Services
	monitorSvc := service.NewMonitorService(monitors, cfg, store)

	// Initialize Controllers
	monitorCtrl := controller.NewMonitorController(monitorSvc)

	// Setup and start the router
	router := api.SetupRouter(monitorCtrl, cfg)
	router.Run() // listens on 0.0.0.0:8080 by default
}
