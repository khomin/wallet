package core

import (
	"context"
	"errors"

	"tracker/internal/db/models"

	"github.com/google/uuid"
)

var ErrWalletNotFound = errors.New("not found")

type WalletRepository interface {
	ListWallets(ctx context.Context, userID string) ([]models.Wallet, error)
	CreateWallet(ctx context.Context, userID string, chain string, address string, label string) (*models.Wallet, error)
	DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error
	GetWallet(ctx context.Context, userID string, id uuid.UUID) (*models.Wallet, error)
}

type WalletService struct {
	walletRepo WalletRepository
}

func NewWalletService(walletRepo WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

func (s *WalletService) ListWallets(ctx context.Context, userID string) ([]models.Wallet, error) {
	return s.walletRepo.ListWallets(ctx, userID)
}

func (s *WalletService) AddWallet(ctx context.Context, userID string, chain string, address string, label string) (*models.Wallet, error) {
	// request.ID = uuid.New()
	// request.CreatedAt = time.Now().UTC()
	// request.UpdatedAt = request.CreatedAt
	wallet, err := s.walletRepo.CreateWallet(ctx, userID, chain, address, label)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	return s.walletRepo.DeleteWallet(ctx, userID, id)
}
