package core

import (
	"context"
	"time"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	List(ctx context.Context, userID string) ([]domain.WalletBalance, error)
	Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletBalance, error)
	Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error)
	Update(ctx context.Context, userID string, id uuid.UUID, req UpdateWallet) (*domain.Wallet, error)
	Delete(ctx context.Context, userID string, id uuid.UUID) error
	ListForSync(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Wallet, error)

	UpdateBalanceSnapshot(ctx context.Context, userID string, id uuid.UUID, snapshot BalanceSnapshot) error
	GetBalanceSnapshot(ctx context.Context, userID string, id uuid.UUID, filter BalanceSnapshotFilter) ([]domain.WalletBalanceSnapshot, error)
}

type BalanceSnapshot struct {
	Crypto float64
	USD    float64
	Time   time.Time
}

type BalanceSnapshotFilter struct {
	From  time.Time
	To    time.Time
	Limit int
}

type UpdateWallet struct {
	Label  string
	Notify bool
}
