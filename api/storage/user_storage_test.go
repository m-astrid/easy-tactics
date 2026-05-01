package storage

import (
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	_ "github.com/mattn/go-sqlite3"
)

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			telegram_id BIGINT UNIQUE NOT NULL,
			username TEXT,
			full_name TEXT,
			role TEXT DEFAULT 'fighter',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return db
}

func TestUserStorage_Create(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	user := &domain.User{
		TelegramID: 123456,
		Username:   "testuser",
		FullName:   "Test User",
		Role:       domain.RoleFighter,
	}

	err := store.Create(user)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	if user.ID == 0 {
		t.Error("Expected user ID to be set")
	}
}

func TestUserStorage_GetByTelegramID(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	store.Create(&domain.User{
		TelegramID: 123456,
		Username:   "testuser",
		FullName:   "Test User",
		Role:       domain.RoleAdmin,
	})

	found, err := store.GetByTelegramID(123456)
	if err != nil {
		t.Fatalf("GetByTelegramID() failed: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user")
	}

	if found.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", found.Username)
	}
}

func TestUserStorage_GetByTelegramID_NotFound(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	found, err := store.GetByTelegramID(999999)
	if err != nil {
		t.Fatalf("GetByTelegramID() failed: %v", err)
	}

	if found != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestUserStorage_Update(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	user := &domain.User{
		TelegramID: 123456,
		Username:   "oldname",
		FullName:   "Old Name",
		Role:       domain.RoleFighter,
	}
	store.Create(user)

	user.Username = "newname"
	user.Role = domain.RoleAdmin

	err := store.Update(user)
	if err != nil {
		t.Fatalf("Update() failed: %v", err)
	}

	found, _ := store.GetByTelegramID(123456)
	if found.Username != "newname" || found.Role != domain.RoleAdmin {
		t.Errorf("Update did not change fields")
	}
}

func TestUserStorage_Delete(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	user := &domain.User{
		TelegramID: 123456,
		Role:       domain.RoleFighter,
	}
	store.Create(user)

	err := store.Delete(user.ID)
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	found, _ := store.GetByTelegramID(123456)
	if found != nil {
		t.Error("Expected user to be deleted")
	}
}

func TestUserStorage_List(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	store := NewUserStorage(db)

	store.Create(&domain.User{TelegramID: 111, Role: domain.RoleFighter})
	store.Create(&domain.User{TelegramID: 222, Role: domain.RoleAdmin})
	store.Create(&domain.User{TelegramID: 333, Role: domain.RoleCoach})

	users, err := store.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(users) != 3 {
		t.Errorf("Expected 3 users, got %d", len(users))
	}
}
