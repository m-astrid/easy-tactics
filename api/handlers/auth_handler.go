package handlers

import (
	"context"

	"github.com/easy-tactics/api/domain"
	"github.com/easy-tactics/api/storage"
)

type AuthHandler struct {
	userStore *storage.UserStorage
}

func NewAuthHandler(userStore *storage.UserStorage) *AuthHandler {
	return &AuthHandler{userStore: userStore}
}

type AddUserRequest struct {
	TelegramId int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"`
}

type RemoveUserRequest struct {
	TelegramId int64 `json:"telegram_id"`
}

type GetUserRequest struct {
	TelegramId int64 `json:"telegram_id"`
}

type ListUsersRequest struct {
	RoleFilter string `json:"role_filter"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

type UpdateUserRoleRequest struct {
	TelegramId int64  `json:"telegram_id"`
	NewRole    string `json:"new_role"`
}

type CheckAccessRequest struct {
	TelegramId   int64  `json:"telegram_id"`
	RequiredRole string `json:"required_role"`
}

type AccessResponse struct {
	Allowed bool   `json:"allowed"`
	Reason  string `json:"reason"`
}

type UserResponse struct {
	TelegramId int64  `json:"telegram_id"`
	Username   string `json:"username"`
	FullName   string `json:"full_name"`
	Role       string `json:"role"`
}

type UserListResponse struct {
	Users []*UserResponse `json:"users"`
	Total int32           `json:"total"`
}

func (h *AuthHandler) AddUser(ctx context.Context, req *AddUserRequest) (*UserResponse, error) {
	role := domain.UserRole(req.Role)
	if !role.IsValid() {
		role = domain.RoleFighter
	}

	user := &domain.User{
		TelegramID: req.TelegramId,
		Username:   req.Username,
		FullName:   req.FullName,
		Role:       role,
	}

	if err := h.userStore.Create(user); err != nil {
		return nil, err
	}

	return &UserResponse{
		TelegramId: user.TelegramID,
		Username:   user.Username,
		FullName:   user.FullName,
		Role:       string(user.Role),
	}, nil
}

func (h *AuthHandler) GetUser(ctx context.Context, req *GetUserRequest) (*UserResponse, error) {
	user, err := h.userStore.GetByTelegramID(req.TelegramId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	return &UserResponse{
		TelegramId: user.TelegramID,
		Username:   user.Username,
		FullName:   user.FullName,
		Role:       string(user.Role),
	}, nil
}

func (h *AuthHandler) UpdateUserRole(ctx context.Context, req *UpdateUserRoleRequest) (*UserResponse, error) {
	user, err := h.userStore.GetByTelegramID(req.TelegramId)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, nil
	}

	user.Role = domain.UserRole(req.NewRole)
	if err := h.userStore.Update(user); err != nil {
		return nil, err
	}

	return &UserResponse{
		TelegramId: user.TelegramID,
		Username:   user.Username,
		FullName:   user.FullName,
		Role:       string(user.Role),
	}, nil
}

func (h *AuthHandler) CheckAccess(ctx context.Context, req *CheckAccessRequest) (*AccessResponse, error) {
	user, err := h.userStore.GetByTelegramID(req.TelegramId)
	if err != nil {
		return &AccessResponse{Allowed: false, Reason: "DB error"}, err
	}

	if user == nil {
		return &AccessResponse{Allowed: true}, nil
	}

	if user.IsBlocked() {
		return &AccessResponse{Allowed: false, Reason: "Пользователь заблокирован"}, nil
	}

	return &AccessResponse{Allowed: true}, nil
}

func (h *AuthHandler) ListUsers(ctx context.Context, req *ListUsersRequest) (*UserListResponse, error) {
	users, err := h.userStore.List()
	if err != nil {
		return nil, err
	}

	var filtered []*domain.User
	if req.RoleFilter != "" {
		for _, u := range users {
			if string(u.Role) == req.RoleFilter {
				filtered = append(filtered, u)
			}
		}
	} else {
		filtered = users
	}

	offset := int(req.Offset)
	limit := int(req.Limit)
	if limit == 0 {
		limit = 10
	}

	if offset > len(filtered) {
		offset = len(filtered)
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	result := filtered[offset:end]

	responses := make([]*UserResponse, len(result))
	for i, u := range result {
		responses[i] = &UserResponse{
			TelegramId: u.TelegramID,
			Username:   u.Username,
			FullName:   u.FullName,
			Role:       string(u.Role),
		}
	}

	return &UserListResponse{
		Users: responses,
		Total: int32(len(filtered)),
	}, nil
}

func (h *AuthHandler) RemoveUser(ctx context.Context, req *RemoveUserRequest) (bool, error) {
	user, err := h.userStore.GetByTelegramID(req.TelegramId)
	if err != nil {
		return false, err
	}

	if user == nil {
		return false, nil
	}

	if err := h.userStore.Delete(user.ID); err != nil {
		return false, err
	}

	return true, nil
}
