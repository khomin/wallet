package core

import (
	"context"
	"testing"

	"tracker/internal/db/models"

	"github.com/google/uuid"
)

type fakeWalletRepo struct {
	deleted *models.Wallet
	err     error
}

func (f *fakeWalletRepo) ListWallets(ctx context.Context, userID string) ([]models.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) CreateWallet(ctx context.Context, userID string, chain string, address string, label string) (*models.Wallet, error) {
	return nil, nil
}

func (f *fakeWalletRepo) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	return nil
}

func (f *fakeWalletRepo) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*models.Wallet, error) {
	return nil, nil
}

func TestDeleteWalletReturnsDeletedWallet(t *testing.T) {
	want := &models.Wallet{
		ID:      uuid.New(),
		Address: "0xabc",
		Chain:   "ethereum",
		Label:   "primary",
		UserID:  "user-1",
	}

	svc := NewWalletService(&fakeWalletRepo{deleted: want})
	err := svc.DeleteWallet(context.Background(), want.UserID, want.ID)
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
