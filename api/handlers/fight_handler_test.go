package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
	_ "github.com/mattn/go-sqlite3"
)

func setupFightTestDB(t *testing.T) *sql.DB {
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

func TestFightHandler_GetFighterFights(t *testing.T) {
	db := setupFightTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "fighter-1", Name: "Tournament"})

	fighterStore.CreateFight(&domain.Fight{
		UUID: "f1", FighterUUID: "fighter-1", TournamentUUID: "t1",
		OpponentName: "Petr", ScoreWin: 15, ScoreLose: 12, Round: "Finals",
	})
	fighterStore.CreateFight(&domain.Fight{
		UUID: "f2", FighterUUID: "fighter-1", TournamentUUID: "t1",
		OpponentName: "Alex", ScoreWin: 10, ScoreLose: 15, Round: "Semi",
	})

	handler := NewFightHandler(fighterStore)

	resp, err := handler.GetFighterFights(context.Background(), &GetFighterFightsRequest{
		FighterUUID: "fighter-1",
		Limit:       10,
	})

	if err != nil {
		t.Fatalf("GetFighterFights() failed: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("Expected 2 fights, got %d", resp.Total)
	}

	if len(resp.Fights) != 2 {
		t.Errorf("Expected 2 fights in list, got %d", len(resp.Fights))
	}
}

func TestFightHandler_GetFighterFights_ByTournament(t *testing.T) {
	db := setupFightTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	fighterStore.Create(&domain.Fighter{UUID: "fighter-2", Slug: "petr", FullName: "Petr"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "fighter-1", Name: "Tournament"})

	fighterStore.CreateFight(&domain.Fight{
		UUID: "f1", FighterUUID: "fighter-1", TournamentUUID: "t1",
		OpponentName: "Petr", ScoreWin: 15, ScoreLose: 12,
	})
	fighterStore.CreateFight(&domain.Fight{
		UUID: "f2", FighterUUID: "fighter-2", TournamentUUID: "t1",
		OpponentName: "Ivan", ScoreWin: 12, ScoreLose: 15,
	})

	handler := NewFightHandler(fighterStore)

	resp, err := handler.GetFighterFights(context.Background(), &GetFighterFightsRequest{
		FighterUUID:    "fighter-1",
		TournamentUUID: "t1",
	})

	if err != nil {
		t.Fatalf("GetFighterFights() failed: %v", err)
	}

	if resp.Total != 1 {
		t.Errorf("Expected 1 fight, got %d", resp.Total)
	}
}

func TestFightHandler_GetFight(t *testing.T) {
	db := setupFightTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	fighterStore.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	fighterStore.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "fighter-1", Name: "Tournament"})

	fighterStore.CreateFight(&domain.Fight{
		UUID: "f1", FighterUUID: "fighter-1", TournamentUUID: "t1",
		OpponentName: "Petr", ScoreWin: 15, ScoreLose: 12, Round: "Finals",
	})

	handler := NewFightHandler(fighterStore)

	resp, err := handler.GetFight(context.Background(), "f1")

	if err != nil {
		t.Fatalf("GetFight() failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected fight, got nil")
	}

	if resp.OpponentName != "Petr" {
		t.Errorf("Expected opponent 'Petr', got '%s'", resp.OpponentName)
	}

	if resp.ScoreWin != 15 {
		t.Errorf("Expected score_win 15, got %d", resp.ScoreWin)
	}

	if resp.Round != "Finals" {
		t.Errorf("Expected round 'Finals', got '%s'", resp.Round)
	}
}

func TestFightHandler_GetFight_NotFound(t *testing.T) {
	db := setupFightTestDB(t)
	defer db.Close()

	fighterStore := storage.NewFighterStorage(db)
	handler := NewFightHandler(fighterStore)

	resp, err := handler.GetFight(context.Background(), "non-existent")

	if err != nil {
		t.Fatalf("GetFight() failed: %v", err)
	}

	if resp != nil {
		t.Error("Expected nil for non-existent fight")
	}
}
