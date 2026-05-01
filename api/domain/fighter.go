package domain

import (
	"fmt"
	"time"
)

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

func (f *Fighter) Validate() error {
	if f.UUID == "" {
		return ErrValidationRequired
	}
	if f.FullName == "" {
		return ErrValidationRequired
	}
	if f.Slug == "" {
		return ErrValidationRequired
	}
	return nil
}

func (t *Tournament) Validate() error {
	if t.UUID == "" {
		return ErrValidationRequired
	}
	if t.Name == "" {
		return ErrValidationRequired
	}
	if t.FighterUUID == "" {
		return ErrValidationRequired
	}
	return nil
}

func (f *Fight) Validate() error {
	if f.UUID == "" {
		return ErrValidationRequired
	}
	if f.FighterUUID == "" {
		return ErrValidationRequired
	}
	if f.TournamentUUID == "" {
		return ErrValidationRequired
	}
	if f.OpponentName == "" {
		return ErrValidationRequired
	}
	return nil
}

var ErrValidationRequired = fmt.Errorf("validation error: required field missing")
