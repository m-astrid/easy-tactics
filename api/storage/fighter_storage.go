package storage

import (
	"database/sql"

	"github.com/easy-tactics/api/domain"
)

type FighterStorage struct {
	db *sql.DB
}

func NewFighterStorage(db *sql.DB) *FighterStorage {
	return &FighterStorage{db: db}
}

func (s *FighterStorage) Create(fighter *domain.Fighter) error {
	result, err := s.db.Exec(`
		INSERT INTO fighters (uuid, slug, full_name, city, club, hemagon_url)
		VALUES (?, ?, ?, ?, ?, ?)
	`, fighter.UUID, fighter.Slug, fighter.FullName, fighter.City, fighter.Club, fighter.HemagonURL)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	fighter.ID = id
	return nil
}

func (s *FighterStorage) GetByUUID(uuid string) (*domain.Fighter, error) {
	row := s.db.QueryRow(`
		SELECT id, uuid, slug, full_name, city, club, hemagon_url, created_at, updated_at
		FROM fighters WHERE uuid = ?
	`, uuid)

	var f domain.Fighter
	err := row.Scan(&f.ID, &f.UUID, &f.Slug, &f.FullName, &f.City, &f.Club, &f.HemagonURL, &f.CreatedAt, &f.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (s *FighterStorage) Search(query string) ([]*domain.Fighter, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, slug, full_name, city, club, hemagon_url, created_at, updated_at
		FROM fighters WHERE full_name LIKE ?
	`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fighters []*domain.Fighter
	for rows.Next() {
		var f domain.Fighter
		if err := rows.Scan(&f.ID, &f.UUID, &f.Slug, &f.FullName, &f.City, &f.Club, &f.HemagonURL, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fighters = append(fighters, &f)
	}

	return fighters, nil
}

func (s *FighterStorage) Update(fighter *domain.Fighter) error {
	_, err := s.db.Exec(`
		UPDATE fighters SET slug = ?, full_name = ?, city = ?, club = ?, hemagon_url = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, fighter.Slug, fighter.FullName, fighter.City, fighter.Club, fighter.HemagonURL, fighter.ID)
	return err
}

func (s *FighterStorage) Delete(id int64) error {
	_, err := s.db.Exec("DELETE FROM fighters WHERE id = ?", id)
	return err
}

func (s *FighterStorage) List() ([]*domain.Fighter, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, slug, full_name, city, club, hemagon_url, created_at, updated_at
		FROM fighters
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fighters []*domain.Fighter
	for rows.Next() {
		var f domain.Fighter
		if err := rows.Scan(&f.ID, &f.UUID, &f.Slug, &f.FullName, &f.City, &f.Club, &f.HemagonURL, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fighters = append(fighters, &f)
	}

	return fighters, nil
}

func (s *FighterStorage) CreateTournament(tournament *domain.Tournament) error {
	result, err := s.db.Exec(`
		INSERT INTO tournaments (uuid, fighter_uuid, name, city, country, start_date, hemagon_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, tournament.UUID, tournament.FighterUUID, tournament.Name, tournament.City, tournament.Country, tournament.StartDate, tournament.HemagonURL)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	tournament.ID = id
	return nil
}

func (s *FighterStorage) GetTournamentsByFighterUUID(fighterUUID string) ([]*domain.Tournament, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, fighter_uuid, name, city, country, start_date, hemagon_url, created_at
		FROM tournaments WHERE fighter_uuid = ?
	`, fighterUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tournaments []*domain.Tournament
	for rows.Next() {
		var t domain.Tournament
		if err := rows.Scan(&t.ID, &t.UUID, &t.FighterUUID, &t.Name, &t.City, &t.Country, &t.StartDate, &t.HemagonURL, &t.CreatedAt); err != nil {
			return nil, err
		}
		tournaments = append(tournaments, &t)
	}

	return tournaments, nil
}

func (s *FighterStorage) CreateFight(fight *domain.Fight) error {
	result, err := s.db.Exec(`
		INSERT INTO fights (uuid, fighter_uuid, tournament_uuid, opponent_uuid, opponent_name, score_win, score_lose, round, fight_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, fight.UUID, fight.FighterUUID, fight.TournamentUUID, fight.OpponentUUID, fight.OpponentName, fight.ScoreWin, fight.ScoreLose, fight.Round, fight.FightDate)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	fight.ID = id
	return nil
}

func (s *FighterStorage) GetFightsByFighterUUID(fighterUUID string) ([]*domain.Fight, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, fighter_uuid, tournament_uuid, opponent_uuid, opponent_name, score_win, score_lose, round, fight_date, created_at
		FROM fights WHERE fighter_uuid = ?
	`, fighterUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fights []*domain.Fight
	for rows.Next() {
		var f domain.Fight
		if err := rows.Scan(&f.ID, &f.UUID, &f.FighterUUID, &f.TournamentUUID, &f.OpponentUUID, &f.OpponentName, &f.ScoreWin, &f.ScoreLose, &f.Round, &f.FightDate, &f.CreatedAt); err != nil {
			return nil, err
		}
		fights = append(fights, &f)
	}

	return fights, nil
}

func (s *FighterStorage) GetFightsByTournamentUUID(tournamentUUID string) ([]*domain.Fight, error) {
	rows, err := s.db.Query(`
		SELECT id, uuid, fighter_uuid, tournament_uuid, opponent_uuid, opponent_name, score_win, score_lose, round, fight_date, created_at
		FROM fights WHERE tournament_uuid = ?
	`, tournamentUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fights []*domain.Fight
	for rows.Next() {
		var f domain.Fight
		if err := rows.Scan(&f.ID, &f.UUID, &f.FighterUUID, &f.TournamentUUID, &f.OpponentUUID, &f.OpponentName, &f.ScoreWin, &f.ScoreLose, &f.Round, &f.FightDate, &f.CreatedAt); err != nil {
			return nil, err
		}
		fights = append(fights, &f)
	}

	return fights, nil
}
