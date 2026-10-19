package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"footballincentivelab/internal/domain"
	"time"
)

func (s *SQLiteStore) GetPlayer(ctx context.Context, id int64) (*domain.Player, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, name, primary_role, secondary_roles, birthdate, club, category, created_at, updated_at FROM players WHERE id = ?", id)
	var p domain.Player
	var secondaryRolesJSON string
	err := row.Scan(&p.ID, &p.Name, &p.PrimaryRole, &secondaryRolesJSON, &p.Birthdate, &p.Club, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get player %d: %w", id, err)
	}
	if err := json.Unmarshal([]byte(secondaryRolesJSON), &p.SecondaryRoles); err != nil {
		return nil, fmt.Errorf("unmarshal secondary roles: %w", err)
	}
	return &p, nil
}

func (s *SQLiteStore) ListPlayers(ctx context.Context) ([]domain.Player, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, name, primary_role, secondary_roles, birthdate, club, category, created_at, updated_at FROM players ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list players: %w", err)
	}
	defer rows.Close()

	var players []domain.Player
	for rows.Next() {
		var p domain.Player
		var secondaryRolesJSON string
		if err := rows.Scan(&p.ID, &p.Name, &p.PrimaryRole, &secondaryRolesJSON, &p.Birthdate, &p.Club, &p.Category, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan player: %w", err)
		}
		if err := json.Unmarshal([]byte(secondaryRolesJSON), &p.SecondaryRoles); err != nil {
			return nil, fmt.Errorf("unmarshal secondary roles: %w", err)
		}
		players = append(players, p)
	}
	return players, rows.Err()
}

func (s *SQLiteStore) CreatePlayer(ctx context.Context, p domain.Player) (int64, error) {
	secondaryRolesJSON, err := json.Marshal(p.SecondaryRoles)
	if err != nil {
		return 0, fmt.Errorf("marshal secondary roles: %w", err)
	}
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO players (name, primary_role, secondary_roles, birthdate, club, category, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		p.Name, p.PrimaryRole, string(secondaryRolesJSON), p.Birthdate, p.Club, p.Category, now, now)
	if err != nil {
		return 0, fmt.Errorf("create player: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) UpdatePlayer(ctx context.Context, p domain.Player) error {
	secondaryRolesJSON, err := json.Marshal(p.SecondaryRoles)
	if err != nil {
		return fmt.Errorf("marshal secondary roles: %w", err)
	}
	_, err = s.db.ExecContext(ctx,
		"UPDATE players SET name=?, primary_role=?, secondary_roles=?, birthdate=?, club=?, category=?, updated_at=? WHERE id=?",
		p.Name, p.PrimaryRole, string(secondaryRolesJSON), p.Birthdate, p.Club, p.Category, time.Now(), p.ID)
	return err
}

func (s *SQLiteStore) DeletePlayer(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM players WHERE id=?", id)
	return err
}
