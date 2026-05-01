package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// MigrationRunner handles database migrations
type MigrationRunner struct {
	db         *sql.DB
	migrations string
}

// NewMigrationRunner creates a new migration runner
func NewMigrationRunner(db *sql.DB, migrationsPath string) *MigrationRunner {
	return &MigrationRunner{
		db:         db,
		migrations: migrationsPath,
	}
}

// RunMigrations executes all pending migrations
func (r *MigrationRunner) RunMigrations() error {
	// Get list of migration files
	files, err := filepath.Glob(filepath.Join(r.migrations, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort by name (ensures order)
	sort.Strings(files)

	log.Printf("Found %d migration files", len(files))

	for _, file := range files {
		if err := r.runMigration(file); err != nil {
			return fmt.Errorf("migration %s failed: %w", file, err)
		}
	}

	return nil
}

// runMigration executes a single migration file
func (r *MigrationRunner) runMigration(file string) error {
	// Check if migration was already applied
	migrationName := filepath.Base(file)
	row := r.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", migrationName)
	var count int
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("failed to check migration status: %w", err)
	}

	if count > 0 {
		log.Printf("Skipping migration: %s (already applied)", migrationName)
		return nil
	}

	// Read and execute migration
	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	if _, err := r.db.Exec(string(content)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record successful migration
	if _, err := r.db.Exec("INSERT INTO schema_migrations (name, applied_at) VALUES (?, datetime('now'))", migrationName); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	log.Printf("Applied migration: %s", migrationName)
	return nil
}

// CreateSchemaMigrationsTable creates the migrations tracking table
func CreateSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			name TEXT PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}
