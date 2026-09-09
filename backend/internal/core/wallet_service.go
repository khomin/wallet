package core

import (
	"context"
	"fmt"
	"time"

	"tracker/internal/core/demo"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type WalletService struct {
	walletRepo        WalletRepository
	walletDemo        *demo.DemoWallets
	priceService      *PriceService
	blockchainService *BlockchainService
	tokenRegistry     *TokenRegistry
	userRepo          UserRepo
}

type WalletDeps struct {
	WalletRepo        WalletRepository
	WalletDemo        *demo.DemoWallets
	PriceService      *PriceService
	UserRepo          UserRepo
	BlockchainService *BlockchainService
	TokenRegistry     *TokenRegistry
}

func NewWalletService(deps WalletDeps) *WalletService {
	return &WalletService{
		walletRepo:        deps.WalletRepo,
		walletDemo:        deps.WalletDemo,
		priceService:      deps.PriceService,
		userRepo:          deps.UserRepo,
		blockchainService: deps.BlockchainService,
		tokenRegistry:     deps.TokenRegistry,
	}
}

func (s *WalletService) GetWallet(ctx context.Context, user *domain.User, id uuid.UUID) (*domain.UserWalletBalance, error) {
	if user.IsDemo {
		return s.walletDemo.GetWallet(id)
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.GetWalletByUser(ctx, user.ID, id)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) ListWallets(ctx context.Context, user *domain.User) ([]domain.UserWalletBalance, error) {
	if user.IsDemo {
		return s.walletDemo.GetWallets(), nil
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return nil, err
	}
	wallets, err := s.walletRepo.ListWalletsByUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	return wallets, nil

}

func (s *WalletService) CreateWallet(ctx context.Context, user *domain.User, chain string, address string, symbol string, label string) error {
	if chain == "" || symbol == "" || address == "" {
		return fmt.Errorf("invalid arguments")
	}
	if user.IsDemo {
		return domain.ErrNotAllowedInDemoMode
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return nil
	}
	if err := s.blockchainService.ValidateAddress(chain, address, symbol); err != nil {
		return err
	}
	_, err := s.walletRepo.Create(ctx, user.ID, chain, address, symbol, label)
	if err != nil {
		return err
	}
	return nil
}

func (s *WalletService) UpdateWallet(ctx context.Context, user *domain.User, id uuid.UUID, req UpdateWallet) (*domain.UserWallet, error) {
	if user.IsDemo {
		return nil, domain.ErrNotAllowedInDemoMode
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.Update(ctx, user.ID, id, req)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, user *domain.User, id uuid.UUID) error {
	if user.IsDemo {
		return domain.ErrNotAllowedInDemoMode
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return err
	}
	return s.walletRepo.Delete(ctx, user.ID, id)
}

func (s *WalletService) FetchBalance(ctx context.Context, wallet domain.Wallet) (*domain.WalletBalanceSnapshot, error) {
	priceSymbol := wallet.Symbol
	if wallet.Chain == wallet.Symbol {
		priceSymbol = wallet.Chain
	}
	token, err := s.blockchainService.tokenRegistry.GetByChainAndSymbol(wallet.Chain, priceSymbol)
	if err != nil {
		return nil, fmt.Errorf("seems like unsupported token %s", priceSymbol)
	}
	price, err := s.priceService.GetPrice(ctx, token.Symbol)
	if err != nil {
		return nil, fmt.Errorf("getting price for %s: %w", priceSymbol, err)
	}
	balance, err := s.blockchainService.GetBalance(ctx, wallet.Chain, wallet.Address, wallet.Symbol)
	if err != nil {
		logrus.Warnf("failed to pull balance for %s on %s: %v", wallet.Address, wallet.Chain, err)
		return nil, err
	}
	return &domain.WalletBalanceSnapshot{
		Time:       time.Now(),
		Balance:    balance.Balance,
		BalanceUSD: balance.Balance * price.CurrentPrice,
	}, nil
}

func (s *WalletService) ListBalanceSnapshots(ctx context.Context, user *domain.User, id uuid.UUID, filter BalanceSnapshotFilter) ([]domain.WalletBalanceSnapshot, error) {
	if user.IsDemo {
		return s.walletDemo.GetWalletBalanceSnapshot(id, filter.From, filter.To)
	}
	if err := s.userRepo.EnsureExists(ctx, user); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.ListBalanceSnapshots(ctx, id, filter)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}
