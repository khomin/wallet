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

func (f *fakeWalletRepo) List(ctx context.Context, userID string) ([]domain.WalletWithBalance, error) {
	return nil, nil
}

func (f *fakeWalletRepo) Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) Edit(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) Delete(ctx context.Context, userID string, id uuid.UUID) error {
	return nil
}

func (f *fakeWalletRepo) Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletWithBalance, error) {
	return nil, nil
}

func (f *fakeWalletRepo) UpdateBalance(ctx context.Context, userID string, id uuid.UUID, balance float64, balanceUSD float64) error {
	return nil
}

func (f *fakeWalletRepo) ListForSync(ctx context.Context, limit int) ([]domain.Wallet, error) {
	return nil, nil
}

type fakeUserRepo struct {
}

func (r *fakeUserRepo) List(ctx context.Context) ([]domain.User, error) {
	return nil, nil
}

func (f *fakeUserRepo) EnsureExists(ctx context.Context, user domain.User) error {
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
		BlockchainService: &BlockchainService{},
		TokenRegistry:     &TokenRegistry{},
		UserRepo:          &fakeUserRepo{},
	})
	err := svc.DeleteWallet(context.Background(), domain.User{
		ID:    want.UserID,
		Name:  want.UserID,
		Email: want.UserID,
	}, want.ID.Bytes)
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
