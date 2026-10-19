-- +goose Up
CREATE TABLE IF NOT EXISTS contracts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    player_id INTEGER NOT NULL,
    season TEXT NOT NULL,
    base_salary INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    level TEXT NOT NULL DEFAULT '',
    max_bonus_ratio REAL NOT NULL DEFAULT 1.0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (player_id) REFERENCES players(id) ON DELETE CASCADE
);

CREATE INDEX idx_contracts_player_id ON contracts(player_id);
CREATE INDEX idx_contracts_season ON contracts(season);

-- +goose Down
DROP TABLE IF EXISTS contracts;
