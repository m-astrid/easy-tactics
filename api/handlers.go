package main

import (
	"database/sql"
	"time"
)

// Add a new fighter to the database
func addFighter(db *sql.DB, uuid, slug, fullName, city, club string) (int64, error) {
	result, err := db.Exec(`
		INSERT INTO fighters (uuid, slug, full_name, city, club, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, uuid, slug, fullName, city, club, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// Search fighters by name
func searchFighters(db *sql.DB, query string) ([]Fighter, error) {
	rows, err := db.Query(`
		SELECT id, uuid, slug, full_name, city, club, hemagon_url, created_at, updated_at
		FROM fighters
		WHERE full_name LIKE ? OR slug LIKE ?
	`, "%"+query+"%", "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fighters []Fighter
	for rows.Next() {
		var f Fighter
		var hemagonURL, club sql.NullString
		err := rows.Scan(&f.ID, &f.UUID, &f.Slug, &f.FullName, &f.City, &club, &hemagonURL, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if hemagonURL.Valid {
			f.HemagonURL = hemagonURL.String
		}
		if club.Valid {
			f.Club = club.String
		}
		fighters = append(fighters, f)
	}
	return fighters, nil
}

// Get fighter by UUID
func getFighterByUUID(db *sql.DB, uuid string) (*Fighter, error) {
	var f Fighter
	var hemagonURL, club sql.NullString
	err := db.QueryRow(`
		SELECT id, uuid, slug, full_name, city, club, hemagon_url, created_at, updated_at
		FROM fighters WHERE uuid = ?
	`, uuid).Scan(&f.ID, &f.UUID, &f.Slug, &f.FullName, &f.City, &club, &hemagonURL, &f.CreatedAt, &f.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if hemagonURL.Valid {
		f.HemagonURL = hemagonURL.String
	}
	if club.Valid {
		f.Club = club.String
	}
	return &f, nil
}

// Fighter represents a fencer
type Fighter struct {
	ID         int64
	UUID       string
	Slug       string
	FullName   string
	City       string
	Club       string
	HemagonURL string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
