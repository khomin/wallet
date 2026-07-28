package dto

import (
	"tracker/internal/core"

	"github.com/google/uuid"
)

type WalletsResponse struct {
	Wallet                 []WalletResponse `json:"wallet"`
	Total                  int              `json:"total"`
	TotalAccountBalanceUsd float64          `json:"total_balance_usd"`
}

type WalletResponse struct {
	ID               string  `json:"id"`
	Address          string  `json:"address"`
	Chain            string  `json:"chain"`
	TokenSymbol      string  `json:"token_symbol"`
	Label            string  `json:"label"`
	BalanceCrypto    float64 `json:"balance_crypto"`
	BalanceUsd       float64 `json:"balance_usd"`
	Change24hPercent float64 `json:"change_24h_percent"`
	HasError         bool    `json:"has_error,omitempty"`
	ErrorMsg         string  `json:"error_msg,omitempty"`
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

func ToWalletResponse(in *core.WalletPortfolio) WalletResponse {
	return WalletResponse{
		ID:               in.Wallet.ID,
		Address:          in.Wallet.Address,
		Chain:            in.Wallet.Chain,
		TokenSymbol:      in.Wallet.Symbol,
		Label:            in.Wallet.Label,
		Change24hPercent: in.Price.PriceChangePercentage_24h,
		BalanceCrypto:    in.Balance,
		BalanceUsd:       in.BalanceUSD,
		HasError:         in.HasError,
		ErrorMsg:         in.ErrorMsg,
	}
}

func ToWalletResponses(in []core.WalletPortfolio) WalletsResponse {
	wallets := make([]WalletResponse, len(in))
	var total float64
	for i, wallet := range in {
		wallets[i] = ToWalletResponse(&wallet)
		total += wallet.BalanceUSD
	}
	return WalletsResponse{
		Total:                  len(wallets),
		Wallet:                 wallets,
		TotalAccountBalanceUsd: total,
	}
}
