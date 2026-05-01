package main

import (
	"database/sql"
	"testing"
)

// TestSearchFighters_EmptyDB returns no results
func TestSearchFighters_EmptyDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	initSchema(db)

	fighters, err := searchFighters(db, "Ivan")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(fighters) != 0 {
		t.Errorf("Expected 0 fighters, got %d", len(fighters))
	}
}

// TestSearchFighters_WithData returns matching fighters
func TestSearchFighters_WithData(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}
	defer db.Close()

	initSchema(db)

	// Add test data
	addFighter(db, "test-uuid-1", "ivan-petrov-msk", "Иван Петров", "Москва", "Спартак")
	addFighter(db, "test-uuid-2", "ivan-ivanov-spb", "Иван Иванов", "Санкт-Петербург", "Звезда")

	fighters, err := searchFighters(db, "Петров")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(fighters) != 1 {
		t.Errorf("Expected 1 fighter, got %d", len(fighters))
	}
	if fighters[0].FullName != "Иван Петров" {
		t.Errorf("Expected 'Иван Петров', got '%s'", fighters[0].FullName)
	}
}

// TestSearchFighters_PartialMatch works
func TestSearchFighters_PartialMatch(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	initSchema(db)

	addFighter(db, "uuid-1", "petr-sidorov-moskow", "Петр Сидоров", "Москва", "Динамо")

	// Search by last name
	fighters, _ := searchFighters(db, "Сидоров")
	if len(fighters) != 1 {
		t.Errorf("Expected 1 fighter, got %d", len(fighters))
	}
}

// TestGetFighterByUUID_NotFound returns nil
func TestGetFighterByUUID_NotFound(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	initSchema(db)

	fighter, err := getFighterByUUID(db, "non-existent-uuid")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if fighter != nil {
		t.Errorf("Expected nil fighter, got %v", fighter)
	}
}

// TestGetFighterByUUID_Found returns fighter
func TestGetFighterByUUID_Found(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	defer db.Close()
	initSchema(db)

	addFighter(db, "uuid-123", "ivan-petrov-msk", "Иван Петров", "Москва", "Спартак")

	fighter, err := getFighterByUUID(db, "uuid-123")
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
	if fighter == nil {
		t.Fatal("Expected fighter, got nil")
	}
	if fighter.FullName != "Иван Петров" {
		t.Errorf("Expected 'Иван Петров', got '%s'", fighter.FullName)
	}
}
