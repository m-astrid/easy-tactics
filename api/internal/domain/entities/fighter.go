package domain

import "time"

// Fighter represents a fencer in the system
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

// NewFighter creates a new fighter with generated UUID and slug
func NewFighter(fullName, city, club, hemagonURL string) *Fighter {
	return &Fighter{
		UUID:       generateUUID(),
		Slug:       generateSlug(fullName, city),
		FullName:   fullName,
		City:       city,
		Club:       club,
		HemagonURL: hemagonURL,
	}
}

// Tournament represents a competition
type Tournament struct {
	ID          int64
	UUID        string
	FighterUUID string
	Name        string
	City        string
	Country     string
	StartDate   string
	HemagonURL  string
	CreatedAt   time.Time
}

// Fight represents a single bout in a tournament
type Fight struct {
	ID             int64
	UUID           string
	FighterUUID    string
	TournamentUUID string
	OpponentUUID   string
	OpponentName   string
	ScoreWin       int
	ScoreLose      int
	Round          string
	FightDate      string
	CreatedAt      time.Time
}
