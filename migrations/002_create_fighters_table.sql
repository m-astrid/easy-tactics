-- Migration: 002_create_fighters_table.sql
-- Description: Fighters table for storing fencer profiles
-- Created: 2026-05-01

CREATE TABLE IF NOT EXISTS fighters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uuid TEXT UNIQUE NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    city TEXT,
    club TEXT,
    hemagon_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_fighters_uuid ON fighters(uuid);
CREATE INDEX IF NOT EXISTS idx_fighters_slug ON fighters(slug);
CREATE INDEX IF NOT EXISTS idx_fighters_full_name ON fighters(full_name);