package storage

import (
	"database/sql"

	"github.com/easy-tactics/api/domain"
)

type UserStorage struct {
	db *sql.DB
}

func NewUserStorage(db *sql.DB) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(user *domain.User) error {
	result, err := s.db.Exec(`
		INSERT INTO users (telegram_id, username, full_name, role)
		VALUES (?, ?, ?, ?)
	`, user.TelegramID, user.Username, user.FullName, user.Role)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = id
	return nil
}

func (s *UserStorage) GetByTelegramID(telegramID int64) (*domain.User, error) {
	row := s.db.QueryRow(`
		SELECT id, telegram_id, username, full_name, role, created_at, updated_at
		FROM users WHERE telegram_id = ?
	`, telegramID)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TelegramID,
		&user.Username,
		&user.FullName,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *UserStorage) Update(user *domain.User) error {
	_, err := s.db.Exec(`
		UPDATE users SET username = ?, full_name = ?, role = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, user.Username, user.FullName, user.Role, user.ID)
	return err
}

func (s *UserStorage) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}

func (s *UserStorage) List() ([]*domain.User, error) {
	rows, err := s.db.Query(`
		SELECT id, telegram_id, username, full_name, role, created_at, updated_at
		FROM users
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.Username,
			&user.FullName,
			&user.Role,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}
