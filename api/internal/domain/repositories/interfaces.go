package domain

// UserRepository defines the interface for user data access
type UserRepository interface {
	FindByTelegramID(telegramID int64) (*User, error)
	FindByRole(role UserRole) ([]User, error)
	Save(user *User) error
	Update(user *User) error
	Delete(telegramID int64) error
	CountByRole(role UserRole) (int, error)
}

// FighterRepository defines the interface for fighter data access
type FighterRepository interface {
	FindByUUID(uuid string) (*Fighter, error)
	FindBySlug(slug string) (*Fighter, error)
	Search(query string) ([]Fighter, error)
	Save(fighter *Fighter) error
	Update(fighter *Fighter) error
	Delete(uuid string) error
}

// TournamentRepository defines the interface for tournament data access
type TournamentRepository interface {
	FindByUUID(uuid string) (*Tournament, error)
	FindByFighter(fighterUUID string) ([]Tournament, error)
	Save(tournament *Tournament) error
}

// FightRepository defines the interface for fight data access
type FightRepository interface {
	FindByUUID(uuid string) (*Fight, error)
	FindByFighter(fighterUUID string) ([]Fight, error)
	FindByTournament(tournamentUUID string) ([]Fight, error)
	Save(fight *Fight) error
}
