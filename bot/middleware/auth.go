package middleware

import (
	"context"
	"fmt"
	"strconv"

	"github.com/easy-tactics/bot/handlers"
)

type AuthMiddleware struct {
	apiClient       *handlers.APIClient
	ownerTelegramID string
}

func NewAuthMiddleware(apiClient *handlers.APIClient, ownerTelegramID string) *AuthMiddleware {
	return &AuthMiddleware{
		apiClient:       apiClient,
		ownerTelegramID: ownerTelegramID,
	}
}

func (m *AuthMiddleware) GetUserRole(telegramID int64) string {
	if m.ownerTelegramID != "" {
		ownerID, err := strconv.ParseInt(m.ownerTelegramID, 10, 64)
		if err == nil && ownerID == telegramID {
			return "owner"
		}
	}

	if m.apiClient == nil {
		return "fighter"
	}

	user, err := m.apiClient.GetUser(context.Background(), telegramID)
	if err != nil || user == nil {
		return "fighter"
	}

	return user.Role
}

func (m *AuthMiddleware) CheckAccess(telegramID int64, requiredRole string) (bool, string) {
	role := m.GetUserRole(telegramID)

	roleHierarchy := map[string]int{
		"owner":   5,
		"admin":   4,
		"coach":   3,
		"fighter": 2,
		"blocked": 0,
	}

	userLevel := roleHierarchy[role]
	requiredLevel := roleHierarchy[requiredRole]

	if userLevel >= requiredLevel {
		return true, ""
	}

	return false, fmt.Sprintf("Требуется роль: %s", requiredRole)
}

func (m *AuthMiddleware) RequireRole(roles ...string) func(telegramID int64) bool {
	return func(telegramID int64) bool {
		userRole := m.GetUserRole(telegramID)
		for _, role := range roles {
			if userRole == role {
				return true
			}
		}
		return false
	}
}
