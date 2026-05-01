package handlers

import (
	"context"
	"database/sql"
	"testing"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
	_ "github.com/mattn/go-sqlite3"
)

func setupAuthTestDB(t *testing.T) *sql.DB {
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

func TestAuthHandler_AddUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	handler := NewAuthHandler(userStore)

	resp, err := handler.AddUser(context.Background(), &AddUserRequest{
		TelegramId: 123456,
		Username:   "testuser",
		FullName:   "Test User",
		Role:       "fighter",
	})

	if err != nil {
		t.Fatalf("AddUser() failed: %v", err)
	}

	if resp.TelegramId != 123456 {
		t.Errorf("Expected telegram_id 123456, got %d", resp.TelegramId)
	}
}

func TestAuthHandler_GetUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	userStore.Create(&domain.User{
		TelegramID: 123456,
		Username:   "testuser",
		FullName:   "Test User",
		Role:       domain.RoleAdmin,
	})

	handler := NewAuthHandler(userStore)

	resp, err := handler.GetUser(context.Background(), &GetUserRequest{
		TelegramId: 123456,
	})

	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if resp.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", resp.Username)
	}
}

func TestAuthHandler_GetUser_NotFound(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	handler := NewAuthHandler(userStore)

	resp, err := handler.GetUser(context.Background(), &GetUserRequest{
		TelegramId: 999999,
	})

	if err != nil {
		t.Fatalf("GetUser() failed: %v", err)
	}

	if resp != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestAuthHandler_UpdateUserRole(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	userStore.Create(&domain.User{
		TelegramID: 123456,
		Role:       domain.RoleFighter,
	})

	handler := NewAuthHandler(userStore)

	resp, err := handler.UpdateUserRole(context.Background(), &UpdateUserRoleRequest{
		TelegramId: 123456,
		NewRole:    "admin",
	})

	if err != nil {
		t.Fatalf("UpdateUserRole() failed: %v", err)
	}

	if resp.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", resp.Role)
	}
}

func TestAuthHandler_CheckAccess_BlockedUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	userStore.Create(&domain.User{
		TelegramID: 123456,
		Role:       domain.RoleBlocked,
	})

	handler := NewAuthHandler(userStore)

	resp, err := handler.CheckAccess(context.Background(), &CheckAccessRequest{
		TelegramId:   123456,
		RequiredRole: "fighter",
	})

	if err != nil {
		t.Fatalf("CheckAccess() failed: %v", err)
	}

	if resp.Allowed {
		t.Error("Expected blocked user to be denied access")
	}
}

func TestAuthHandler_CheckAccess_NewUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	handler := NewAuthHandler(userStore)

	resp, err := handler.CheckAccess(context.Background(), &CheckAccessRequest{
		TelegramId:   456789,
		RequiredRole: "fighter",
	})

	if err != nil {
		t.Fatalf("CheckAccess() failed: %v", err)
	}

	if !resp.Allowed {
		t.Error("Expected new user to be allowed access")
	}
}

func TestAuthHandler_ListUsers(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	userStore.Create(&domain.User{TelegramID: 111, Role: domain.RoleFighter})
	userStore.Create(&domain.User{TelegramID: 222, Role: domain.RoleAdmin})
	userStore.Create(&domain.User{TelegramID: 333, Role: domain.RoleCoach})

	handler := NewAuthHandler(userStore)

	resp, err := handler.ListUsers(context.Background(), &ListUsersRequest{
		Limit: 10,
	})

	if err != nil {
		t.Fatalf("ListUsers() failed: %v", err)
	}

	if resp.Total != 3 {
		t.Errorf("Expected 3 users, got %d", resp.Total)
	}
}

func TestAuthHandler_RemoveUser(t *testing.T) {
	db := setupAuthTestDB(t)
	defer db.Close()

	userStore := storage.NewUserStorage(db)
	userStore.Create(&domain.User{
		TelegramID: 123456,
		Role:       domain.RoleFighter,
	})

	handler := NewAuthHandler(userStore)

	_, err := handler.RemoveUser(context.Background(), &RemoveUserRequest{
		TelegramId: 123456,
	})

	if err != nil {
		t.Fatalf("RemoveUser() failed: %v", err)
	}

	found, _ := userStore.GetByTelegramID(123456)
	if found != nil {
		t.Error("Expected user to be removed")
	}
}
