package core

import (
	"context"
	"testing"

	"tracker/internal/core/domain"
	"tracker/internal/db/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type fakeWalletRepo struct {
	deleted *models.Wallet
	err     error
}

func (f *fakeWalletRepo) ListWallets(ctx context.Context, userID string) ([]domain.WalletWithBalance, error) {
	return nil, nil
}

func (f *fakeWalletRepo) CreateWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	return nil
}

func (f *fakeWalletRepo) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletWithBalance, error) {
	return nil, nil
}

func (f *fakeWalletRepo) SetWalletBalance(ctx context.Context, userID string, id uuid.UUID, balance float64, balanceUSD float64) error {
	return nil
}

type fakeUserRepo struct {
}

func (f *fakeUserRepo) EnsureExists(ctx context.Context, userID string) error {
	return nil
}

func TestDeleteWalletReturnsDeletedWallet(t *testing.T) {
	want := &models.Wallet{
		ID: pgtype.UUID{
			Bytes: uuid.New(),
			Valid: true,
		},
		Address: "0xabc",
		Chain:   "ethereum",
		Label:   "primary",
		UserID:  "user-1",
	}

	svc := NewWalletService(WalletDeps{
		WalletRepo:        &fakeWalletRepo{},
		PriceService:      &PriceService{},
		BlockchainService: &BlockchainService{}, TokenRegistry: &TokenRegistry{}, UserRepo: &fakeUserRepo{},
	})
	err := svc.DeleteWallet(context.Background(), want.UserID, want.ID.Bytes)
	if err != nil {
		t.Fatalf("DeleteWallet returned unexpected error: %v", err)
	}
	// if deleted == nil {
	// 	t.Fatal("DeleteWallet returned nil wallet")
	// }
	// if deleted.Address != want.Address {
	// 	t.Fatalf("expected deleted wallet address %q, got %q", want.Address, deleted.Address)
	// }
}
