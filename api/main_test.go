package main

import (
	"database/sql"
	"testing"
)

// Test that the schema initializes correctly
func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Initialize schema first
	if err := initSchema(db); err != nil {
		t.Fatalf("Failed to init schema: %v", err)
	}

	// Test that users table exists and can be queried
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query users table: %v", err)
	}

	// Test that fighters table exists
	err = db.QueryRow("SELECT COUNT(*) FROM fighters").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query fighters table: %v", err)
	}
}

// Test that owner creation works
func TestInitOwner(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	// Create schema
	initSchema(db)

	// Try to create owner
	_, err = db.Exec(`
		INSERT INTO users (telegram_id, username, full_name, role)
		VALUES (123456789, 'test_user', 'Test User', 'owner')
	`)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Verify owner exists
	var role string
	err = db.QueryRow("SELECT role FROM users WHERE telegram_id = ?", 123456789).Scan(&role)
	if err != nil {
		t.Fatalf("Failed to query owner: %v", err)
	}
	if role != "owner" {
		t.Errorf("Expected role 'owner', got '%s'", role)
	}
}

// Test that blocked user is properly identified
func TestBlockedUser(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	initSchema(db)

	// Insert blocked user
	db.Exec(`INSERT INTO users (telegram_id, role) VALUES (999, 'blocked')`)

	// Query for blocked status
	var role string
	err = db.QueryRow("SELECT role FROM users WHERE telegram_id = ?", 999).Scan(&role)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}
	if role != "blocked" {
		t.Errorf("Expected 'blocked', got '%s'", role)
	}
}
