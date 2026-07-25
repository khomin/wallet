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

type WalletPortfolioItem struct {
	Wallet     models.Wallet
	Price      models.CoinPrice
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
	priceService      *PriceService
	blockchainService *BlockchainService
	tokenRegistry     *TokenRegistry
}

func NewWalletService(
	walletRepo WalletRepository,
	priceService *PriceService,
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
	// TODO: add validation
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

func (s *WalletService) EditWallet(ctx context.Context, userID string, id uuid.UUID, label string) (*WalletPortfolioItem, error) {
	wallet, err := s.walletRepo.EditWallet(ctx, userID, id, label)
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

// a list of tokens
func (s *WalletService) GetAssetsByMartket(ctx context.Context) ([]entity.Asset, error) {
	assets := []entity.Asset{}

	// max := 50

	// get top coins
	coins, err := s.priceService.GetCoins(ctx)
	if err != nil {
		return nil, err
	}
	// build assets from top coins copying imageURL from COIN
	for _, coin := range coins {
		asset, found := s.tokenRegistry.GetByChainAndSymbol(coin.Symbol, coin.Symbol)
		if !found {
			continue
		}
		// for index := range assets {
		// asset = assets[index]
		asset.LogoURL = coin.ImageURL
		assets[index] = asset
		// }
		assets = append(assets, assets...)
	}
	return assets, err
}

func (s *WalletService) SearchAssets(ctx context.Context, text string) ([]entity.Asset, error) {
	// assets := []entity.Asset{}
	assets := s.tokenRegistry.GetAssetsByText(text)
	// assets = s.assingIcon(ctx, assets)
	return assets, nil
}

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

func (s *WalletService) getWalletPortfolio(ctx context.Context, wallet *models.Wallet) (*WalletPortfolioItem, error) {
	priceSymbol := wallet.Symbol
	if wallet.Chain == wallet.Symbol {
		priceSymbol = wallet.Chain
	}
	token, found := s.blockchainService.tokenRegistry.GetByChainAndSymbol(wallet.Chain, priceSymbol)
	if !found {
		return &WalletPortfolioItem{
			Wallet: *wallet,
			Price:  models.CoinPrice{},
		}, fmt.Errorf("seems like unsupported token %s", priceSymbol)
	}
	// price, err := s.priceService.GetPrice(ctx, priceSymbol)
	price, err := s.priceService.GetPrice(ctx, token.Symbol)
	if err != nil {
		return &WalletPortfolioItem{
			Wallet: *wallet,
			Price:  models.CoinPrice{},
		}, fmt.Errorf("getting price for %s: %w", priceSymbol, err)
	}
	item := &WalletPortfolioItem{
		Wallet: *wallet,
		Price:  *price,
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
