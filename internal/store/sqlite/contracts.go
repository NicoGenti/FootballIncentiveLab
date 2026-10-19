package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"footballincentivelab/internal/domain"
	"time"
)

func (s *SQLiteStore) GetContract(ctx context.Context, id int64) (*domain.Contract, error) {
	row := s.db.QueryRowContext(ctx,
		"SELECT id, player_id, season, base_salary, start_date, end_date, level, max_bonus_ratio, created_at, updated_at FROM contracts WHERE id = ?", id)
	var c domain.Contract
	err := row.Scan(&c.ID, &c.PlayerID, &c.Season, &c.BaseSalary, &c.StartDate, &c.EndDate, &c.Level, &c.MaxBonusRatio, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get contract %d: %w", id, err)
	}
	return &c, nil
}

func (s *SQLiteStore) ListContracts(ctx context.Context) ([]domain.Contract, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, player_id, season, base_salary, start_date, end_date, level, max_bonus_ratio, created_at, updated_at FROM contracts ORDER BY start_date DESC")
	if err != nil {
		return nil, fmt.Errorf("list contracts: %w", err)
	}
	defer rows.Close()

	var contracts []domain.Contract
	for rows.Next() {
		var c domain.Contract
		if err := rows.Scan(&c.ID, &c.PlayerID, &c.Season, &c.BaseSalary, &c.StartDate, &c.EndDate, &c.Level, &c.MaxBonusRatio, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

func (s *SQLiteStore) ListContractsByPlayer(ctx context.Context, playerID int64) ([]domain.Contract, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, player_id, season, base_salary, start_date, end_date, level, max_bonus_ratio, created_at, updated_at FROM contracts WHERE player_id = ? ORDER BY start_date DESC", playerID)
	if err != nil {
		return nil, fmt.Errorf("list contracts for player %d: %w", playerID, err)
	}
	defer rows.Close()

	var contracts []domain.Contract
	for rows.Next() {
		var c domain.Contract
		if err := rows.Scan(&c.ID, &c.PlayerID, &c.Season, &c.BaseSalary, &c.StartDate, &c.EndDate, &c.Level, &c.MaxBonusRatio, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan contract: %w", err)
		}
		contracts = append(contracts, c)
	}
	return contracts, rows.Err()
}

func (s *SQLiteStore) CreateContract(ctx context.Context, c domain.Contract) (int64, error) {
	now := time.Now()
	result, err := s.db.ExecContext(ctx,
		"INSERT INTO contracts (player_id, season, base_salary, start_date, end_date, level, max_bonus_ratio, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.PlayerID, c.Season, c.BaseSalary, c.StartDate, c.EndDate, c.Level, c.MaxBonusRatio, now, now)
	if err != nil {
		return 0, fmt.Errorf("create contract: %w", err)
	}
	return result.LastInsertId()
}

func (s *SQLiteStore) UpdateContract(ctx context.Context, c domain.Contract) error {
	_, err := s.db.ExecContext(ctx,
		"UPDATE contracts SET player_id=?, season=?, base_salary=?, start_date=?, end_date=?, level=?, max_bonus_ratio=?, updated_at=? WHERE id=?",
		c.PlayerID, c.Season, c.BaseSalary, c.StartDate, c.EndDate, c.Level, c.MaxBonusRatio, time.Now(), c.ID)
	return err
}

func (s *SQLiteStore) DeleteContract(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM contracts WHERE id=?", id)
	return err
}
