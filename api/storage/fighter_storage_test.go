package storage

import (
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupFighterDB(t *testing.T) *sql.DB {
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

func TestFighterStorage_Create(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	fighter := &domain.Fighter{
		UUID:       "test-uuid-123",
		Slug:       "ivan-petrov",
		FullName:   "Ivan Petrov",
		City:       "Moscow",
		Club:       "Sword Club",
		HemagonURL: "https://hemagon.com/fighters/ivan-petrov",
	}

	err := store.Create(fighter)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if fighter.ID == 0 {
		t.Error("Expected fighter ID to be set")
	}
}

func TestFighterStorage_GetByUUID(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{
		UUID:     "test-uuid-123",
		Slug:     "ivan-petrov",
		FullName: "Ivan Petrov",
	})

	found, err := store.GetByUUID("test-uuid-123")
	if err != nil {
		t.Fatalf("GetByUUID() failed: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find fighter")
	}

	if found.FullName != "Ivan Petrov" {
		t.Errorf("Expected name 'Ivan Petrov', got '%s'", found.FullName)
	}
}

func TestFighterStorage_GetByUUID_NotFound(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	found, err := store.GetByUUID("non-existent")
	if err != nil {
		t.Fatalf("GetByUUID() failed: %v", err)
	}

	if found != nil {
		t.Error("Expected nil for non-existent fighter")
	}
}

func TestFighterStorage_Search(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "1", Slug: "ivan", FullName: "Ivan"})
	store.Create(&domain.Fighter{UUID: "2", Slug: "petr", FullName: "Petr"})
	store.Create(&domain.Fighter{UUID: "3", Slug: "alex", FullName: "Alex"})

	results, err := store.Search("Ivan")
	if err != nil {
		t.Fatalf("Search() failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}
}

func TestFighterStorage_Update(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	f := &domain.Fighter{
		UUID:     "test-uuid",
		Slug:     "ivan",
		FullName: "Ivan",
		City:     "Moscow",
	}
	store.Create(f)

	f.City = "SPB"
	f.Club = "New Club"

	err := store.Update(f)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	found, _ := store.GetByUUID("test-uuid")
	if found.City != "SPB" || found.Club != "New Club" {
		t.Error("Update did not change fields")
	}
}

func TestFighterStorage_Delete(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	f := &domain.Fighter{UUID: "test-uuid", Slug: "ivan", FullName: "Ivan"}
	store.Create(f)

	err := store.Delete(f.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	found, _ := store.GetByUUID("test-uuid")
	if found != nil {
		t.Error("Expected fighter to be deleted")
	}
}

func TestFighterStorage_List(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "1", Slug: "a", FullName: "A"})
	store.Create(&domain.Fighter{UUID: "2", Slug: "b", FullName: "B"})
	store.Create(&domain.Fighter{UUID: "3", Slug: "c", FullName: "C"})

	fighters, err := store.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(fighters) != 3 {
		t.Errorf("Expected 3 fighters, got %d", len(fighters))
	}
}

func TestTournamentStorage_Create(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})

	tournament := &domain.Tournament{
		UUID:        "tourn-1",
		FighterUUID: "fighter-1",
		Name:        "Moscow Cup 2024",
		City:        "Moscow",
		Country:     "Russia",
		StartDate:   "2024-03-15",
	}

	err := store.CreateTournament(tournament)
	if err != nil {
		t.Fatalf("CreateTournament() failed: %v", err)
	}

	if tournament.ID == 0 {
		t.Error("Expected tournament ID to be set")
	}
}

func TestTournamentStorage_GetByFighterUUID(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})

	store.CreateTournament(&domain.Tournament{
		UUID:        "t1",
		FighterUUID: "fighter-1",
		Name:        "Tournament 1",
	})
	store.CreateTournament(&domain.Tournament{
		UUID:        "t2",
		FighterUUID: "fighter-1",
		Name:        "Tournament 2",
	})

	tournaments, err := store.GetTournamentsByFighterUUID("fighter-1")
	if err != nil {
		t.Fatalf("GetTournamentsByFighterUUID() failed: %v", err)
	}

	if len(tournaments) != 2 {
		t.Errorf("Expected 2 tournaments, got %d", len(tournaments))
	}
}

func TestFightStorage_Create(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "fighter-1", Slug: "ivan", FullName: "Ivan"})
	store.CreateTournament(&domain.Tournament{UUID: "tourn-1", FighterUUID: "fighter-1", Name: "Tournament"})

	fight := &domain.Fight{
		UUID:           "fight-1",
		FighterUUID:    "fighter-1",
		TournamentUUID: "tourn-1",
		OpponentUUID:   "opponent-1",
		OpponentName:   "Petr",
		ScoreWin:       15,
		ScoreLose:      12,
		Round:          "Finals",
		FightDate:      "2024-03-16",
	}

	err := store.CreateFight(fight)
	if err != nil {
		t.Fatalf("CreateFight() failed: %v", err)
	}

	if fight.ID == 0 {
		t.Error("Expected fight ID to be set")
	}
}

func TestFightStorage_GetByFighterUUID(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "f1", Slug: "ivan", FullName: "Ivan"})
	store.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "f1", Name: "T"})

	store.CreateFight(&domain.Fight{UUID: "fight1", FighterUUID: "f1", TournamentUUID: "t1", OpponentName: "A"})
	store.CreateFight(&domain.Fight{UUID: "fight2", FighterUUID: "f1", TournamentUUID: "t1", OpponentName: "B"})

	fights, err := store.GetFightsByFighterUUID("f1")
	if err != nil {
		t.Fatalf("GetFightsByFighterUUID() failed: %v", err)
	}

	if len(fights) != 2 {
		t.Errorf("Expected 2 fights, got %d", len(fights))
	}
}

func TestFightStorage_GetByTournamentUUID(t *testing.T) {
	db := setupFighterDB(t)
	defer db.Close()

	store := NewFighterStorage(db)

	store.Create(&domain.Fighter{UUID: "f1", Slug: "ivan", FullName: "Ivan"})
	store.CreateTournament(&domain.Tournament{UUID: "t1", FighterUUID: "f1", Name: "T"})

	store.CreateFight(&domain.Fight{UUID: "f1", FighterUUID: "f1", TournamentUUID: "t1", OpponentName: "A"})
	store.CreateFight(&domain.Fight{UUID: "f2", FighterUUID: "f1", TournamentUUID: "t1", OpponentName: "B"})

	fights, err := store.GetFightsByTournamentUUID("t1")
	if err != nil {
		t.Fatalf("GetFightsByTournamentUUID() failed: %v", err)
	}

	if len(fights) != 2 {
		t.Errorf("Expected 2 fights, got %d", len(fights))
	}
}
