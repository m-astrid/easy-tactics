package domain

import "testing"

// Test UserRole validation
func TestUserRole_IsValid(t *testing.T) {
	tests := []struct {
		role UserRole
		want bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleCoach, true},
		{RoleFighter, true},
		{RoleBlocked, true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			if got := tt.role.IsValid(); got != tt.want {
				t.Errorf("UserRole.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test User CanManageUsers
func TestUser_CanManageUsers(t *testing.T) {
	tests := []struct {
		role UserRole
		want bool
	}{
		{RoleOwner, true},
		{RoleAdmin, true},
		{RoleCoach, false},
		{RoleFighter, false},
		{RoleBlocked, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.CanManageUsers(); got != tt.want {
				t.Errorf("User.CanManageUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test User IsBlocked
func TestUser_IsBlocked(t *testing.T) {
	tests := []struct {
		role UserRole
		want bool
	}{
		{RoleOwner, false},
		{RoleAdmin, false},
		{RoleCoach, false},
		{RoleFighter, false},
		{RoleBlocked, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.IsBlocked(); got != tt.want {
				t.Errorf("User.IsBlocked() = %v, want %v", got, tt.want)
			}
		})
	}
}
