package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

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

type WalletRepository interface {
	List(ctx context.Context, userID string) ([]domain.WalletWithBalance, error)
	Create(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*domain.Wallet, error)
	Edit(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error)
	Delete(ctx context.Context, userID string, id uuid.UUID) error
	Get(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletWithBalance, error)
	SetBalance(ctx context.Context, userID string, id uuid.UUID, balance float64, balanceUSD float64) error
	ListForSync(ctx context.Context, limit int) ([]domain.Wallet, error)
}

type WalletService struct {
	walletRepo        WalletRepository
	priceService      *PriceService
	blockchainService *BlockchainService
	eventPublisher    messaging.Publisher
	tokenRegistry     *TokenRegistry
	userRepo          UserRepo
}

type WalletDeps struct {
	WalletRepo        WalletRepository
	PriceService      *PriceService
	UserRepo          UserRepo
	EventPublisher    messaging.Publisher
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

func (s *WalletService) GetWallet(ctx context.Context, userID string, id uuid.UUID) (*domain.WalletWithBalance, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) ListWallets(ctx context.Context, userID string) ([]domain.WalletWithBalance, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil, err
	}
	wallets, err := s.walletRepo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	return wallets, nil
}

func (s *WalletService) AddWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) error {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil
	}
	if err := s.blockchainService.ValidateAddress(chain, address, symbol); err != nil {
		return err
	}
	wallet, err := s.walletRepo.Create(ctx, userID, chain, address, symbol, label)
	if err != nil {
		return err
	}
	if bytes, err := json.Marshal(domain.WalletCreatedEvent{
		ID:     wallet.ID,
		UserID: userID,
	}); err != nil {
		return nil
	} else {
		if err := s.eventPublisher.Publish(bytes); err != nil {
			return err
		}
	}
	return nil
}

func (s *WalletService) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*domain.Wallet, error) {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.Edit(ctx, userID, id, label)
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	if err := s.userRepo.EnsureExists(ctx, userID); err != nil {
		return err
	}
	return s.walletRepo.Delete(ctx, userID, id)
}

func (s *WalletService) SynchronizeWallet(ctx context.Context, wallet domain.Wallet) error {
	priceSymbol := wallet.Symbol
	if wallet.Chain == wallet.Symbol {
		priceSymbol = wallet.Chain
	}
	token, err := s.blockchainService.tokenRegistry.GetByChainAndSymbol(wallet.Chain, priceSymbol)
	if err != nil {
		return fmt.Errorf("seems like unsupported token %s", priceSymbol)
	}
	price, err := s.priceService.GetPrice(ctx, token.Symbol)
	if err != nil {
		return fmt.Errorf("getting price for %s: %w", priceSymbol, err)
	}
	balance, err := s.blockchainService.GetBalance(ctx, wallet.Chain, wallet.Address, wallet.Symbol)
	if err != nil {
		logrus.Warnf("failed to pull balance for %s on %s: %v", wallet.Address, wallet.Chain, err)
		return nil
	}
	uuid, err := uuid.Parse(wallet.ID)
	if err != nil {
		return err
	}
	err = s.walletRepo.SetBalance(ctx, wallet.UserID, uuid, balance.Balance, balance.Balance*price.CurrentPrice)
	if err != nil {
		return err
	}
	return nil
}

func (s *WalletService) SynchronizeWallets(ctx context.Context) error {
	wallets, err := s.walletRepo.ListForSync(ctx, 1000)
	if err != nil {
		slog.Warn("failed to fetch active wallets for sync", "error", err)
		return err
	}

	for _, w := range wallets {
		// Wait for rate-limiter token before hitting external gRPC/RPC
		if err := s.rateLimiter.Wait(ctx); err != nil {
			return
		}

		go func(wallet Wallet) {
			provider, ok := s.providers[wallet.Chain]
			if !ok {
				return
			}

			balance, err := provider.GetBalance(ctx, wallet.Address)
			if err != nil {
				// Update DB with the error state you showed above!
				_ = s.repo.UpdateBalanceError(ctx, wallet.ID, "Unable to fetch live balance")
				return
			}

			// Success - clear error and set new balance
			_ = s.repo.UpdateBalanceSuccess(ctx, wallet.ID, balance)
		}(w)
	}
	return nil
}
