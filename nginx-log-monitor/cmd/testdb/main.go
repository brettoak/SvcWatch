package main

import (
	"fmt"
	mon "nginx-log-monitor"
	"nginx-log-monitor/internal/model"
	"nginx-log-monitor/internal/storage"
	"time"
)

func main() {
	// Let's test just the storage creation and insertion
	fmt.Println("Testing SQLite Storage...")
	store := storage.NewSqliteStorage("test_nginx_logs.db", true)
	defer store.Close()

	entry := &model.LogEntry{
		RemoteAddr:    "127.0.0.1",
		RemoteUser:    "-",
		TimeLocal:     time.Now(),
		Request:       "GET / HTTP/1.1",
		Status:        200,
		BodyBytesSent: 1024,
		HttpReferer:   "-",
		HttpUserAgent: "TestAgent/1.0",
	}

	err := store.Save(entry)
	if err != nil {
		fmt.Printf("Error saving entry: %v\n", err)
		return
	}

	count := store.GetTotalCount()
	fmt.Printf("Total logs in database: %d\n", count)
	
	// Create a monitor object using NewMonitor, passing a dummy path to test it doesn't crash
	_, err = mon.NewMonitor("../access.log", true)
	if err != nil {
		fmt.Printf("Monitor initialization err: %v\n", err)
	} else {
		fmt.Println("Monitor initialized successfully. SQLite database nginx_logs.db should be created.")
	}
}
