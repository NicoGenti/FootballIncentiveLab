package store

import (
	"context"
	"footballincentivelab/internal/domain"
)

type Store interface {
	// Players
	GetPlayer(ctx context.Context, id int64) (*domain.Player, error)
	ListPlayers(ctx context.Context) ([]domain.Player, error)
	CreatePlayer(ctx context.Context, p domain.Player) (int64, error)
	UpdatePlayer(ctx context.Context, p domain.Player) error
	DeletePlayer(ctx context.Context, id int64) error

	// Contracts
	GetContract(ctx context.Context, id int64) (*domain.Contract, error)
	ListContracts(ctx context.Context) ([]domain.Contract, error)
	ListContractsByPlayer(ctx context.Context, playerID int64) ([]domain.Contract, error)
	CreateContract(ctx context.Context, c domain.Contract) (int64, error)
	UpdateContract(ctx context.Context, c domain.Contract) error
	DeleteContract(ctx context.Context, id int64) error
}
