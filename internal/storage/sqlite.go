package storage

import (
	"database/sql"
	"fmt"
	"log"
	"SvcWatch/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStorage implements SQLite storage.
type SqliteStorage struct {
	db *sql.DB
}

// NewSqliteStorage creates a new SqliteStorage.
func NewSqliteStorage(dbPath string) *SqliteStorage {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open sqlite database: %v", err)
	}

	return &SqliteStorage{
		db: db,
	}
}

// InitTable initializes a mapped log table in the database.
func (s *SqliteStorage) InitTable(tableName string, clearOnStartup bool) {
	if clearOnStartup {
		// Drop table to clear previous data on startup
		dropTableSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, tableName)
		_, err := s.db.Exec(dropTableSQL)
		if err != nil {
			log.Fatalf("Failed to drop table %s: %v", tableName, err)
		}
	}

	// Create table if not exists
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
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
	`, tableName)
	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table %s: %v", tableName, err)
	}
}

// Save saves a log entry to SQLite.
func (s *SqliteStorage) Save(tableName string, entry *model.LogEntry) error {
	insertSQL := fmt.Sprintf(`
	INSERT INTO %s (
		remote_addr, remote_user, time_local, request, status, body_bytes_sent, http_referer, http_user_agent
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
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
		log.Printf("Failed to insert log entry into %s: %v", tableName, err)
		return err
	}
	return nil
}

// GetTotalCount returns the total number of logs from a SQLite table.
func (s *SqliteStorage) GetTotalCount(tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("Failed to get total count from %s: %v", tableName, err)
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
