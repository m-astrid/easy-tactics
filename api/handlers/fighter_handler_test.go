package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
	_ "github.com/mattn/go-sqlite3"
)

func setupFighterTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db.Exec(`
		CREATE TABLE fighters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid VARCHAR(36) NOT NULL UNIQUE,
			slug VARCHAR(255) NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			city VARCHAR(255),
			club VARCHAR(255),
			hemagon_url VARCHAR(512),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE tournaments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid VARCHAR(36) NOT NULL UNIQUE,
			fighter_uuid VARCHAR(36) NOT NULL,
			name VARCHAR(255) NOT NULL,
			city VARCHAR(255),
			country VARCHAR(255),
			start_date DATE,
			hemagon_url VARCHAR(512),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE fights (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			uuid VARCHAR(36) NOT NULL UNIQUE,
			fighter_uuid VARCHAR(36) NOT NULL,
			tournament_uuid VARCHAR(36) NOT NULL,
			opponent_uuid VARCHAR(36),
			opponent_name VARCHAR(255) NOT NULL,
			score_win INTEGER DEFAULT 0,
			score_lose INTEGER DEFAULT 0,
			round VARCHAR(50),
			fight_date DATE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	return db
}

func TestFighterHandler_SearchFighter(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{
		UUID:     "test-uuid-1",
		Slug:     "ivan-petrov",
		FullName: "Ivan Petrov",
		City:     "Moscow",
	})

	handler := NewFighterHandler(fighterStore)

	resp, err := handler.SearchFighter(context.Background(), &SearchFighterRequest{
		Query: "Petrov",
	})

	if err != nil {
		t.Fatalf("SearchFighter() failed: %v", err)
	}

	if resp.Source != Source_SOURCE_LOCAL {
		t.Errorf("Expected SOURCE_LOCAL, got %v", resp.Source)
	}

	if len(resp.Matches) != 1 {
		t.Errorf("Expected 1 match, got %d", len(resp.Matches))
	}
}

func TestFighterHandler_SearchFighter_NotFound(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	handler := NewFighterHandler(fighterStore)

	resp, err := handler.SearchFighter(context.Background(), &SearchFighterRequest{
		Query: "NonExistent",
	})

	if err != nil {
		t.Fatalf("SearchFighter() failed: %v", err)
	}

	if resp.Source != Source_SOURCE_NOT_FOUND {
		t.Errorf("Expected SOURCE_NOT_FOUND, got %v", resp.Source)
	}
}

func TestFighterHandler_GetFighter(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{
		UUID:     "test-uuid-1",
		Slug:     "ivan-petrov",
		FullName: "Ivan Petrov",
		City:     "Moscow",
		Club:     "Sword Club",
	})

	handler := NewFighterHandler(fighterStore)

	resp, err := handler.GetFighter(context.Background(), &GetFighterRequest{
		Uuid: "test-uuid-1",
	})

	if err != nil {
		t.Fatalf("GetFighter() failed: %v", err)
	}

	if resp.FullName != "Ivan Petrov" {
		t.Errorf("Expected full name 'Ivan Petrov', got '%s'", resp.FullName)
	}

	if resp.City != "Moscow" {
		t.Errorf("Expected city 'Moscow', got '%s'", resp.City)
	}
}

func TestFighterHandler_GetFighter_BySlug(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{
		UUID:     "test-uuid-1",
		Slug:     "ivan-petrov",
		FullName: "Ivan Petrov",
	})

	handler := NewFighterHandler(fighterStore)

	resp, err := handler.GetFighter(context.Background(), &GetFighterRequest{
		Slug: "ivan-petrov",
	})

	if err != nil {
		t.Fatalf("GetFighter() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected fighter, got nil")
	}
}

func TestFighterHandler_CreateFighter(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	handler := NewFighterHandler(fighterStore)

	resp, err := handler.CreateFighter(context.Background(), &CreateFighterRequest{
		FullName: "John Doe",
		City:     "New York",
		Club:     "Fencing Club",
	})

	if err != nil {
		t.Fatalf("CreateFighter() failed: %v", err)
	}

	if resp.Uuid == "" {
		t.Error("Expected UUID to be set")
	}

	if resp.Slug == "" {
		t.Error("Expected Slug to be set")
	}

	if resp.FullName != "John Doe" {
		t.Errorf("Expected full name 'John Doe', got '%s'", resp.FullName)
	}
}

func TestFighterHandler_UpdateFighter(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{
		UUID:     "test-uuid-1",
		Slug:     "ivan-petrov",
		FullName: "Ivan Petrov",
		City:     "Moscow",
	})

	handler := NewFighterHandler(fighterStore)

	resp, err := handler.UpdateFighter(context.Background(), &UpdateFighterRequest{
		Uuid:     "test-uuid-1",
		FullName: "Ivan Petrov Jr",
		City:     "SPB",
	})

	if err != nil {
		t.Fatalf("UpdateFighter() failed: %v", err)
	}

	if resp.FullName != "Ivan Petrov Jr" {
		t.Errorf("Expected full name 'Ivan Petrov Jr', got '%s'", resp.FullName)
	}

	if resp.City != "SPB" {
		t.Errorf("Expected city 'SPB', got '%s'", resp.City)
	}
}

func TestFighterHandler_ListFighters(t *testing.T) {
	db := setupFighterTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "1", Slug: "a", FullName: "A"})
	fighterStore.Create(&domain.Fighter{UUID: "2", Slug: "b", FullName: "B"})
	fighterStore.Create(&domain.Fighter{UUID: "3", Slug: "c", FullName: "C"})

	handler := NewFighterHandler(fighterStore)

	resp, err := handler.ListFighters(context.Background(), &ListFightersRequest{
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("ListFighters() failed: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("Expected 3 fighters, got %d", resp.Total)
	}
}
