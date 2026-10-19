package domain

import "time"

type Contract struct {
	ID            int64     `json:"id"`
	PlayerID      int64     `json:"player_id"`
	Season        string    `json:"season"`
	BaseSalary    int64     `json:"base_salary"` // Stored in cents (integer arithmetic)
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	Level         string    `json:"level"`
	MaxBonusRatio float64   `json:"max_bonus_ratio"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
