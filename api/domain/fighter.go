package domain

import "time"

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

// Fight represents a bout
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
