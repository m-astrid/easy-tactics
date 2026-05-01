-- Migration: 003_create_tournaments_table.sql
-- Description: Tournaments table for storing competition data
-- Created: 2026-05-01

CREATE TABLE IF NOT EXISTS tournaments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    fighter_uuid TEXT REFERENCES fighters(uuid) ON DELETE CASCADE,
    name TEXT NOT NULL,
    city TEXT,
    country TEXT,
    start_date DATE,
    hemagon_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tournaments_fighter ON tournaments(fighter_uuid);
CREATE INDEX IF NOT EXISTS idx_tournaments_start_date ON tournaments(start_date);