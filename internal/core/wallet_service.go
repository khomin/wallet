package core

import (
	"context"
	"errors"

	"tracker/internal/db/models"

	"github.com/google/uuid"
)

var (
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrWalletInternalError = errors.New("wallet internal error")
)

type WalletPortfolioItem struct {
	Wallet     models.Wallet
	Price      models.CoinPrice
	Balance    float64
	BalanceUSD float64
}

type WalletRepository interface {
	ListWallets(ctx context.Context, userID string) ([]models.Wallet, error)
	CreateWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*models.Wallet, error)
	DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error
	GetWallet(ctx context.Context, userID string, id uuid.UUID) (*models.Wallet, error)
}

type WalletService struct {
	walletRepo        WalletRepository
	priceService      *PriceService
	blockchainService *BlockchainService
}

func NewWalletService(
	walletRepo WalletRepository,
	priceService *PriceService,
	blockchainService *BlockchainService,
) *WalletService {
	return &WalletService{
		walletRepo:        walletRepo,
		priceService:      priceService,
		blockchainService: blockchainService,
	}
}

func (s *WalletService) ListWallets(ctx context.Context, userID string) ([]WalletPortfolioItem, error) {
	wallets, err := s.walletRepo.ListWallets(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := []WalletPortfolioItem{}
	for _, wallet := range wallets {
		portfolio, err := s.getWalletPortfolio(ctx, &wallet)
		if err != nil {
			continue
		}
		res = append(res, *portfolio)
	}
	return res, nil
}

func (s *WalletService) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*WalletPortfolioItem, error) {
	wallet, err := s.walletRepo.GetWallet(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return s.getWalletPortfolio(ctx, wallet)
}

func (s *WalletService) AddWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*WalletPortfolioItem, error) {
	wallet, err := s.walletRepo.CreateWallet(ctx, userID, chain, address, symbol, label)
	if err != nil {
		return nil, err
	}
	portfolio, err := s.getWalletPortfolio(ctx, wallet)
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	return s.walletRepo.DeleteWallet(ctx, userID, id)
}

func (s *WalletService) getWalletPortfolio(ctx context.Context, wallet *models.Wallet) (*WalletPortfolioItem, error) {
	price, err := s.priceService.GetPrice(ctx, wallet.Symbol)
	if err != nil {
		return nil, errors.New("cannot pull wallet price")
	}
	balance, err := s.blockchainService.GetBalance(ctx, wallet.Chain, wallet.Address)
	if err != nil {
		return nil, errors.New("cannot pull wallet balance")
	}
	return &WalletPortfolioItem{
		Wallet:     *wallet,
		Price:      *price,
		Balance:    balance.Balance,
		BalanceUSD: balance.Balance * price.CurrentPrice,
	}, nil
}
