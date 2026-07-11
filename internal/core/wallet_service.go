package core

import (
	"context"
	"time"

	"tracker/internal/db/models"
	repositories "tracker/internal/db/repo"

	"github.com/google/uuid"
)

type WalletService struct {
	walletRepo *repositories.WalletRepository
}

func NewWalletService(walletRepo *repositories.WalletRepository) *WalletService {
	return &WalletService{walletRepo: walletRepo}
}

func (s *WalletService) ListWallets(ctx context.Context) ([]models.Wallet, error) {
	return s.walletRepo.ListWallets(ctx)
}

func (s *WalletService) AddWallet(ctx context.Context, request models.Wallet) (*models.Wallet, error) {
	request.ID = uuid.New()
	request.CreatedAt = time.Now().UTC()
	request.UpdatedAt = request.CreatedAt
	if err := s.walletRepo.CreateWallet(ctx, request); err != nil {
		return nil, err
	}
	return &request, nil
}
