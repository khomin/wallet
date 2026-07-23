package dto

import (
	"time"

	"tracker/internal/core"

	"github.com/google/uuid"
)

type WalletsResponse struct {
	Wallet                 []WalletResponse `json:"wallet"`
	Total                  int              `json:"total"`
	TotalAccountBalanceUsd float64          `json:"total_balance_usd"`
}

type WalletResponse struct {
	ID               uuid.UUID `json:"id"`
	Address          string    `json:"address"`
	Chain            string    `json:"chain"`
	TokenSymbol      string    `json:"token_symbol"`
	Label            string    `json:"label"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	BalanceCrypto    float64   `json:"balance_crypto"`
	BalanceUsd       float64   `json:"balance_usd"`
	Change24hPercent float64   `json:"change_24h_percent"`
	HasError         bool      `json:"has_error,omitempty"`
	ErrorMsg         string    `json:"error_msg,omitempty"`
}

type CreateWalletRequest struct {
	Chain       string `json:"chain" binding:"required"`
	Address     string `json:"address" binding:"required"`
	TokenSymbol string `json:"token_symbol" binding:"required"`
	Label       string `json:"label,omitempty"`
}

type EditWalletRequest struct {
	ID    uuid.UUID `json:"id" binding:"required"`
	Label string    `json:"label,omitempty"`
}

type EditWalletResponse struct {
	WalletResponse
}

type DeleteWalletRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

type DeleteWalletResponse struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

type GetWalletBalanceResponse struct {
	Chain       string  `json:"chain"`
	TokenSymbol string  `json:"token_symbol"`
	Address     string  `json:"address" `
	Balance     float64 `json:"balance"`
	BalanceUSD  float64 `json:"balance_usd"`
}

type GetWalletBalanceRequest struct {
	ID uuid.UUID `form:"id" json:"id"`
}

func ToWalletResponse(walletPortfolio *core.WalletPortfolioItem) WalletResponse {
	return WalletResponse{
		ID:               walletPortfolio.Wallet.ID,
		Address:          walletPortfolio.Wallet.Address,
		Chain:            walletPortfolio.Wallet.Chain,
		TokenSymbol:      walletPortfolio.Wallet.Symbol,
		Label:            walletPortfolio.Wallet.Label,
		CreatedAt:        walletPortfolio.Wallet.CreatedAt,
		UpdatedAt:        walletPortfolio.Wallet.UpdatedAt,
		Change24hPercent: walletPortfolio.Price.PriceChangePercentage_24h,
		BalanceCrypto:    walletPortfolio.Balance,
		BalanceUsd:       walletPortfolio.BalanceUSD,
		HasError:         walletPortfolio.HasError,
		ErrorMsg:         walletPortfolio.ErrorMsg,
	}
}

func ToWalletResponses(wallets []core.WalletPortfolioItem) WalletsResponse {
	wallets_ := make([]WalletResponse, len(wallets))
	var total float64
	for i, wallet := range wallets {
		wallets_[i] = ToWalletResponse(&wallet)
		total += wallet.BalanceUSD
	}
	return WalletsResponse{
		Total:                  len(wallets_),
		Wallet:                 wallets_,
		TotalAccountBalanceUsd: total,
	}
}

func ToGetWalletBalanceResponse(walletPortfolio *core.WalletPortfolioItem) GetWalletBalanceResponse {
	return GetWalletBalanceResponse{
		Chain:      walletPortfolio.Wallet.Chain,
		Address:    walletPortfolio.Wallet.Address,
		Balance:    walletPortfolio.Balance,
		BalanceUSD: walletPortfolio.BalanceUSD,
	}
}
