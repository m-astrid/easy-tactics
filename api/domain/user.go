package domain

import "time"

// User represents a bot user with role-based access
type User struct {
	ID         int64
	TelegramID int64
	Username   string
	FullName   string
	Role       UserRole
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleOwner   UserRole = "owner"
	RoleAdmin   UserRole = "admin"
	RoleCoach   UserRole = "coach"
	RoleFighter UserRole = "fighter"
	RoleBlocked UserRole = "blocked"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleCoach, RoleFighter, RoleBlocked:
		return true
	}
	return false
}

// CanManageUsers returns true if the user can manage other users
func (u *User) CanManageUsers() bool {
	return u.Role == RoleOwner || u.Role == RoleAdmin
}

// IsBlocked returns true if the user is blocked
func (u *User) IsBlocked() bool {
	return u.Role == RoleBlocked
}
