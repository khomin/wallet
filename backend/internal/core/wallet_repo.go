package core

import (
	"context"
	"time"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
)

type WalletRepository interface {
	ListUsersByWallet(ctx context.Context, id uuid.UUID) ([]domain.UserWallet, error)
	GetWalletByUser(ctx context.Context, userID string, id uuid.UUID) (*domain.UserWalletBalance, error)
	ListWalletsByUser(ctx context.Context, userID string) ([]domain.UserWalletBalance, error)

	Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.UserWallet, error)
	Update(ctx context.Context, userID string, id uuid.UUID, req UpdateWallet) (*domain.UserWallet, error)
	Delete(ctx context.Context, userID string, id uuid.UUID) error

	UpdateBalanceSnapshot(ctx context.Context, id uuid.UUID, snapshot BalanceSnapshot) error
	ListBalanceSnapshots(ctx context.Context, id uuid.UUID, filter BalanceSnapshotFilter) ([]domain.WalletBalanceSnapshot, error)
	GetBalanceSnapshot(ctx context.Context, id uuid.UUID) (*domain.WalletBalanceSnapshot, error)
	ListForSync(ctx context.Context, updatedAt time.Time, limit int) ([]domain.Wallet, error)
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
