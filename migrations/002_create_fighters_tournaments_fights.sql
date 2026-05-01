-- +goose NOFILE
-- Description: Create fighters, tournaments, and fights tables

-- Create fighters table
CREATE TABLE IF NOT EXISTS fighters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    slug VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    city VARCHAR(255),
    club VARCHAR(255),
    hemagon_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fighters_uuid ON fighters(uuid);
CREATE INDEX IF NOT EXISTS idx_fighters_slug ON fighters(slug);

-- Create tournaments table
CREATE TABLE IF NOT EXISTS tournaments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    fighter_uuid VARCHAR(36) NOT NULL REFERENCES fighters(uuid),
    name VARCHAR(255) NOT NULL,
    city VARCHAR(255),
    country VARCHAR(255),
    start_date DATE,
    hemagon_url VARCHAR(512),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tournaments_uuid ON tournaments(uuid);
CREATE INDEX IF NOT EXISTS idx_tournaments_fighter_uuid ON tournaments(fighter_uuid);

-- Create fights table
CREATE TABLE IF NOT EXISTS fights (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid VARCHAR(36) NOT NULL UNIQUE,
    fighter_uuid VARCHAR(36) NOT NULL REFERENCES fighters(uuid),
    tournament_uuid VARCHAR(36) NOT NULL REFERENCES tournaments(uuid),
    opponent_uuid VARCHAR(36),
    opponent_name VARCHAR(255) NOT NULL,
    score_win INTEGER DEFAULT 0,
    score_lose INTEGER DEFAULT 0,
    round VARCHAR(50),
    fight_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fights_uuid ON fights(uuid);
CREATE INDEX IF NOT EXISTS idx_fights_fighter_uuid ON fights(fighter_uuid);
CREATE INDEX IF NOT EXISTS idx_fights_tournament_uuid ON fights(tournament_uuid);