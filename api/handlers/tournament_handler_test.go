package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
	_ "github.com/mattn/go-sqlite3"
)

func setupTournamentTestDB(t *testing.T) *sql.DB {
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

func TestTournamentHandler_GetFighterTournaments(t *testing.T) {
	db := setupTournamentTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "fighter-1", Name: "Tournament 1", City: "Moscow"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t2", FighterUUID: "fighter-1", Name: "Tournament 2", City: "SPB"})

	handler := NewTournamentHandler(fighterStore)

	resp, err := handler.GetFighterTournaments(context.Background(), &GetFighterTournamentsRequest{
		FighterUUID: "fighter-1",
		Limit:       10,
	})

	if err != nil {
		t.Fatalf("GetFighterTournaments() failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("Expected 2 tournaments, got %d", resp.Total)
	}

	if len(resp.Tournaments) != 2 {
		t.Errorf("Expected 2 tournaments in list, got %d", len(resp.Tournaments))
	}
}

func TestTournamentHandler_GetFighterTournaments_Empty(t *testing.T) {
	db := setupTournamentTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})

	handler := NewTournamentHandler(fighterStore)

	resp, err := handler.GetFighterTournaments(context.Background(), &GetFighterTournamentsRequest{
		FighterUUID: "fighter-1",
	})

	if err != nil {
		t.Fatalf("GetFighterTournaments() failed: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("Expected 0 tournaments, got %d", resp.Total)
	}
}

func TestTournamentHandler_GetTournament(t *testing.T) {
	db := setupTournamentTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "fighter-1", Name: "Moscow Cup", City: "Moscow", Country: "Russia"})

	handler := NewTournamentHandler(fighterStore)

	resp, err := handler.GetTournament(context.Background(), "t1")

	if err != nil {
		t.Fatalf("GetTournament() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected tournament, got nil")
	}

	if resp.Name != "Moscow Cup" {
		t.Errorf("Expected name 'Moscow Cup', got '%s'", resp.Name)
	}

	if resp.Country != "Russia" {
		t.Errorf("Expected country 'Russia', got '%s'", resp.Country)
	}
}

func TestTournamentHandler_GetTournament_NotFound(t *testing.T) {
	db := setupTournamentTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	handler := NewTournamentHandler(fighterStore)

	resp, err := handler.GetTournament(context.Background(), "non-existent")

	if err != nil {
		t.Fatalf("GetTournament() failed: %v", err)
	}

	if resp != nil {
		t.Error("Expected nil for non-existent tournament")
	}
}
