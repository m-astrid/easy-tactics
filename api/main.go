package main

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	"github.com/easy-tactics/api/database"
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

func main() {
	cfg := Load()
	log.Printf("Starting API service with DB: %s", cfg.DBPath)

	// Connect to database
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize migrations
	if err := database.CreateSchemaMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	migrationRunner := database.NewMigrationRunner(db, "/app/migrations")
	if err := migrationRunner.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Try to create owner
	if err := initOwner(db, cfg.OwnerTelegramID); err != nil {
		log.Printf("Warning: owner creation failed: %v", err)
	}

	log.Println("API service initialized successfully")
}

// initOwner creates the first owner user if specified in env
func initOwner(db *sql.DB, ownerID string) error {
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
