package database

import (
	"database/sql"
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
	return &MigrationRunner{db: db, migrations: migrationsPath}
}

// RunMigrations executes all pending migrations
func (r *MigrationRunner) RunMigrations() error {
	files, err := filepath.Glob(filepath.Join(r.migrations, "*.sql"))
	if err != nil {
		return err
	}

	sort.Strings(files)
	log.Printf("Found %d migration files", len(files))

	for _, file := range files {
		if err := r.runMigration(file); err != nil {
			return err
		}
	}
	return nil
}

func (r *MigrationRunner) runMigration(file string) error {
	migrationName := filepath.Base(file)
	var count int
	_ = r.db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ?", migrationName).Scan(&count)
	if count > 0 {
		log.Printf("Skipping migration: %s (already applied)", migrationName)
		return nil
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	if _, err := r.db.Exec(string(content)); err != nil {
		return err
	}

	_, _ = r.db.Exec("INSERT INTO schema_migrations (name, applied_at) VALUES (?, datetime('now'))", migrationName)
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
