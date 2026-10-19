-- +goose Up
CREATE TABLE IF NOT EXISTS players (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    primary_role TEXT NOT NULL,
    secondary_roles TEXT NOT NULL DEFAULT '[]',
    birthdate DATE NOT NULL,
    club TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT '',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_players_club ON players(club);
CREATE INDEX idx_players_primary_role ON players(primary_role);

-- +goose Down
DROP TABLE IF EXISTS players;
