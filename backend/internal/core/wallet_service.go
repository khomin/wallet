package core

import (
	"context"
	"errors"
	"fmt"

	"tracker/internal/core/entity"
	"tracker/internal/db/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	ErrWalletNotFound      = errors.New("not found")
	ErrWalletAlreadyExists = errors.New("already exists")
	ErrWalletInternalError = errors.New("internal error")
)

type WalletPortfolio struct {
	Wallet     entity.Wallet
	Price      entity.TokenPrice
	Balance    float64
	BalanceUSD float64
	HasError   bool
	ErrorMsg   string
}

type WalletRepository interface {
	ListWallets(ctx context.Context, userID string) ([]models.Wallet, error)
	CreateWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*models.Wallet, error)
	EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*models.Wallet, error)
	DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error
	GetWallet(ctx context.Context, userID string, id uuid.UUID) (*models.Wallet, error)
}

type WalletService struct {
	walletRepo        WalletRepository
	priceService      *AssetService
	blockchainService *BlockchainService
	tokenRegistry     *TokenRegistry
}

func NewWalletService(
	walletRepo WalletRepository,
	priceService *AssetService,
	blockchainService *BlockchainService,
	tokenRegistry *TokenRegistry,
) *WalletService {
	return &WalletService{
		walletRepo:        walletRepo,
		priceService:      priceService,
		blockchainService: blockchainService,
		tokenRegistry:     tokenRegistry,
	}
}

func (s *WalletService) GetWallet(ctx context.Context, userID string, id uuid.UUID) (WalletPortfolio, error) {
	wallet, err := s.walletRepo.GetWallet(ctx, userID, id)
	if err != nil {
		return WalletPortfolio{}, err
	}
	return s.getWalletPortfolio(ctx, walletDbToEntity(*wallet))
}

func (s *WalletService) ListWallets(ctx context.Context, userID string) ([]WalletPortfolio, error) {
	res := []WalletPortfolio{}
	wallets, err := s.walletRepo.ListWallets(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, wallet := range wallets {
		portfolio, err := s.getWalletPortfolio(ctx, walletDbToEntity(wallet))
		if err != nil {
			continue
		}
		res = append(res, portfolio)
	}
	return res, nil
}

func (s *WalletService) AddWallet(ctx context.Context, userID string, chain string, address string, symbol string, label string) (*WalletPortfolio, error) {
	// TODO: add validation
	wallet, err := s.walletRepo.CreateWallet(ctx, userID, chain, address, symbol, label)
	if err != nil {
		return nil, err
	}
	portfolio, err := s.getWalletPortfolio(ctx, walletDbToEntity(*wallet))
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (s *WalletService) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*WalletPortfolio, error) {
	wallet, err := s.walletRepo.EditWallet(ctx, userID, id, label)
	if err != nil {
		return nil, err
	}
	portfolio, err := s.getWalletPortfolio(ctx, walletDbToEntity(*wallet))
	if err != nil {
		return nil, err
	}
	return &portfolio, nil
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID string, id uuid.UUID) error {
	return s.walletRepo.DeleteWallet(ctx, userID, id)
}

// a list of tokens
// func (s *WalletService) GetAssetsByMartket(ctx context.Context) ([]entity.Asset, error) {
// 	assets := []entity.Asset{}

// 	// max := 50

// 	// get top coins
// 	coins, err := s.priceService.GetCoins(ctx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// build assets from top coins copying imageURL from COIN
// 	for _, coin := range coins {
// 		asset, found := s.tokenRegistry.GetByChainAndSymbol(coin.Symbol, coin.Symbol)
// 		if !found {
// 			continue
// 		}
// 		// for index := range assets {
// 		// asset = assets[index]
// 		asset.LogoURL = coin.ImageURL
// 		assets[index] = asset
// 		// }
// 		assets = append(assets, assets...)
// 	}
// 	return assets, err
// }

// func (s *WalletService) assingIcon(ctx context.Context, in []entity.Asset) []entity.Asset {
// 	assets := []entity.Asset{}
// 	for _, i := range in {
// 		for _, token := range i.Tokens {
// 			coin, err := s.priceService.GetCoin(ctx, token.Symbol)
// 			if err == nil {
// 				token.LogoURL = coin.ImageURL
// 			}
// 		}
// 		// assets = append(assets, )
// 	}
// 	return assets
// }

func (s *WalletService) getWalletPortfolio(ctx context.Context, wallet entity.Wallet) (WalletPortfolio, error) {
	priceSymbol := wallet.Symbol
	if wallet.Chain == wallet.Symbol {
		priceSymbol = wallet.Chain
	}
	primarySymbol, _, found := s.blockchainService.tokenRegistry.GetByChainAndSymbol(wallet.Chain, priceSymbol)
	if !found {
		return WalletPortfolio{
			Wallet: wallet,
			Price:  entity.TokenPrice{},
		}, fmt.Errorf("seems like unsupported token %s", priceSymbol)
	}
	price, err := s.priceService.GetPrice(ctx, primarySymbol)
	if err != nil {
		return WalletPortfolio{
			Wallet: wallet,
			Price:  entity.TokenPrice{},
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

func walletDbToEntity(in models.Wallet) entity.Wallet {
	return entity.Wallet{
		ID:      in.ID.String(),
		Address: in.Address,
		Chain:   in.Chain,
		Label:   in.Label,
		Symbol:  in.Symbol,
		UserID:  in.UserID,
	}
}
