package storage

import (
	"database/sql"
	"log"
	"nginx-log-monitor/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStorage implements SQLite storage.
type SqliteStorage struct {
	db *sql.DB
}

// NewSqliteStorage creates a new SqliteStorage.
func NewSqliteStorage(dbPath string, clearOnStartup bool) *SqliteStorage {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open sqlite database: %v", err)
	}

	if clearOnStartup {
		// Drop table to clear previous data on startup
		dropTableSQL := `DROP TABLE IF EXISTS nginx_logs;`
		_, err = db.Exec(dropTableSQL)
		if err != nil {
			log.Fatalf("Failed to drop table: %v", err)
		}
	}

	// Create table if not exists
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS nginx_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		remote_addr TEXT,
		remote_user TEXT,
		time_local DATETIME,
		request TEXT,
		status INTEGER,
		body_bytes_sent INTEGER,
		http_referer TEXT,
		http_user_agent TEXT
	);
	`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	return &SqliteStorage{
		db: db,
	}
}

// Save saves a log entry to SQLite.
func (s *SqliteStorage) Save(entry *model.LogEntry) error {
	insertSQL := `
	INSERT INTO nginx_logs (
		remote_addr, remote_user, time_local, request, status, body_bytes_sent, http_referer, http_user_agent
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := s.db.Exec(insertSQL,
		entry.RemoteAddr,
		entry.RemoteUser,
		entry.TimeLocal,
		entry.Request,
		entry.Status,
		entry.BodyBytesSent,
		entry.HttpReferer,
		entry.HttpUserAgent,
	)
	if err != nil {
		log.Printf("Failed to insert log entry: %v", err)
		return err
	}
	return nil
}

// GetTotalCount returns the total number of logs from SQLite.
func (s *SqliteStorage) GetTotalCount() int {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM nginx_logs").Scan(&count)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		return 0
	}
	return count
}

// Close closes the database connection.
func (s *SqliteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
