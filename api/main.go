package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Config holds application configuration
type Config struct {
	DBPath          string
	OwnerTelegramID string
}

// Load configuration from environment
func Load() Config {
	return Config{
		DBPath:          getEnv("DB_PATH", "/data/fighters.db"),
		OwnerTelegramID: getEnv("OWNER_TELEGRAM_ID", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// User represents a bot user
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FullName   string
	Role       string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// initOwner creates the first owner user if specified in env
func initOwner(db *sql.DB) error {
	ownerID := os.Getenv("OWNER_TELEGRAM_ID")
	if ownerID == "" {
		log.Println("No OWNER_TELEGRAM_ID specified, skipping owner creation")
		return nil
	}

	telegramID, err := strconv.ParseInt(ownerID, 10, 64)
	if err != nil {
		log.Printf("Invalid OWNER_TELEGRAM_ID: %v", err)
		return err
	}

	// Check if owner already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'owner'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("Owner already exists")
		return nil
	}

	// Create owner
	_, err = db.Exec(`
		INSERT INTO users (telegram_id, username, full_name, role, created_at, updated_at)
		VALUES (?, 'owner', 'Owner', 'owner', datetime('now'), datetime('now'))
	`, telegramID)
	if err != nil {
		log.Printf("Failed to create owner: %v", err)
		return err
	}

	log.Printf("Created owner with telegram_id: %d", telegramID)
	return nil
}

func main() {
	cfg := Load()
	log.Printf("Starting API service with DB: %s", cfg.DBPath)

	// Connect to database
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize schema
	if err := initSchema(db); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Try to create owner
	if err := initOwner(db); err != nil {
		log.Printf("Warning: owner creation failed: %v", err)
	}

	log.Println("API service initialized successfully")
}

func initSchema(db *sql.DB) error {
	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id BIGINT UNIQUE NOT NULL,
			username TEXT,
			full_name TEXT,
			role TEXT DEFAULT 'fighter',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id)`,
		`CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)`,
		`CREATE TABLE IF NOT EXISTS fighters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			full_name TEXT NOT NULL,
			city TEXT,
			club TEXT,
			hemagon_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tournaments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			fighter_uuid TEXT,
			name TEXT NOT NULL,
			city TEXT,
			country TEXT,
			start_date DATE,
			hemagon_url TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS fights (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid TEXT UNIQUE NOT NULL,
			fighter_uuid TEXT,
			tournament_uuid TEXT,
			opponent_uuid TEXT,
			opponent_name TEXT,
			score_win INTEGER,
			score_lose INTEGER,
			round TEXT,
			fight_date DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, schema := range schemas {
		if _, err := db.Exec(schema); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
	}

	log.Println("Database schema initialized")
	return nil
}
