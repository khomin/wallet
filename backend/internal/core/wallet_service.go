package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"tracker/internal/core/domain"
	"tracker/internal/messaging"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrWalletNotFound      = errors.New("not found")
	ErrWalletAlreadyExists = errors.New("already exists")
	ErrWalletInternalError = errors.New("internal error")
)

type WalletPortfolio struct {
	Wallet     domain.Wallet
	Price      domain.TokenPrice
	Balance    float64
	BalanceUSD float64
	HasError   bool
	ErrorMsg   string
}

type WalletRepository interface {
	ListWallets(ctx context.Context, userID string) ([]domain.Wallet, error)
	CreateWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error)
	EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error)
	DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error
	GetWallet(ctx context.Context, userID string, id uuid.UUID) (*domain.Wallet, error)
}

type WalletService struct {
	walletRepo        WalletRepository
	priceService      *PriceService
	blockchainService *BlockchainService
	eventPublisher    *messaging.Publisher
	tokenRegistry     *TokenRegistry
	userRepo          UserRepo
}

type WalletDeps struct {
	WalletRepo        WalletRepository
	PriceService      *PriceService
	UserRepo          UserRepo
	EventPublisher    *messaging.Publisher
	BlockchainService *BlockchainService
	TokenRegistry     *TokenRegistry
}

func NewWalletService(deps WalletDeps) *WalletService {
	return &WalletService{
		walletRepo:        deps.WalletRepo,
		priceService:      deps.PriceService,
		userRepo:          deps.UserRepo,
		blockchainService: deps.BlockchainService,
		tokenRegistry:     deps.TokenRegistry,
		eventPublisher:    deps.EventPublisher,
	}
}

func (s *WalletService) GetWallet(ctx context.Context, userID string, id uuid.UUID) (WalletPortfolio, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return WalletPortfolio{}, err
	}
	wallet, err := s.walletRepo.GetWallet(ctx, userID, id)
	if err != nil {
		return WalletPortfolio{}, err
	}
	return s.FetchPortfolio(ctx, *wallet)
}

func (s *WalletService) ListWallets(ctx context.Context, userID string) ([]WalletPortfolio, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil, err
	}
	wallets, err := s.walletRepo.ListWallets(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := []WalletPortfolio{}
	for _, wallet := range wallets {
		portfolio, err := s.FetchPortfolio(ctx, wallet)
		if err != nil {
			continue
		}
		out = append(out, portfolio)
	}
	return out, nil
}

func (s *WalletService) AddWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) error {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil
	}
	if err := s.blockchainService.ValidateAddress(chain, address, symbol); err != nil {
		return err
	}
	wallet, err := s.walletRepo.CreateWallet(ctx, userID, chain, address, symbol, label)
	if err != nil {
		return err
	}
	event := domain.WalletCreatedEvent{
		ID:     wallet.ID,
		UserID: userID,
	}
	if bytes, err := json.Marshal(event); err != nil {
		return nil
	} else {
		if err := s.eventPublisher.Publish(bytes); err != nil {
			return err
		}
	}
	return nil
}

func (s *WalletService) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*WalletPortfolio, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.EditWallet(ctx, userID, id, label)
	if err != nil {
		return nil, err
	}
	portfolio, err := s.FetchPortfolio(ctx, *wallet)
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return err
	}
	return s.walletRepo.DeleteWallet(ctx, userID, id)
}

func (s *WalletService) FetchPortfolio(ctx context.Context, wallet domain.Wallet) (WalletPortfolio, error) {
	priceSymbol := wallet.Symbol
	if wallet.Chain == wallet.Symbol {
		priceSymbol = wallet.Chain
	}
	token, err := s.blockchainService.tokenRegistry.GetByChainAndSymbol(wallet.Chain, priceSymbol)
	if err != nil {
		return WalletPortfolio{
			Wallet: wallet,
			Price:  domain.TokenPrice{},
		}, fmt.Errorf("seems like unsupported token %s", priceSymbol)
	}
	price, err := s.priceService.GetPrice(ctx, token.Symbol)
	if err != nil {
		return WalletPortfolio{
			Wallet: wallet,
			Price:  domain.TokenPrice{},
		}, fmt.Errorf("getting price for %s: %w", priceSymbol, err)
	}
	item := WalletPortfolio{
		Wallet: wallet,
		Price:  price,
	}
	balance, err := s.blockchainService.GetBalance(ctx, wallet.Chain, wallet.Address, wallet.Symbol)
	if err != nil {
		logrus.Warnf("failed to pull balance for %s on %s: %v", wallet.Address, wallet.Chain, err)
		item.HasError = true
		item.ErrorMsg = "Unable to fetch live balance"
		return item, nil
	}
	item.Balance = balance.Balance
	item.BalanceUSD = balance.Balance * price.CurrentPrice
	return item, nil
}
