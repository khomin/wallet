package dto

import (
	"time"

	"tracker/internal/core"
	"tracker/internal/db/models"

	"github.com/google/uuid"
)

type WalletResponse struct {
	ID        uuid.UUID `json:"id"`
	Address   string    `json:"address"`
	Chain     string    `json:"chain"`
	Label     string    `json:"label"`
	UserID    string    `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateWalletRequest struct {
	Chain   string `json:"chain" binding:"required"`
	Address string `json:"address" binding:"required"`
	Label   string `json:"label,omitempty"`
}

type DeleteWalletResponse struct {
	ID uuid.UUID `json:"id"`
}

type DeleteWalletRequest struct {
	ID uuid.UUID `json:"id"`
}

type GetWalletBalanceResponse struct {
	Chain   string  `json:"chain" binding:"required"`
	Address string  `json:"address" binding:"required"`
	Balance float64 `json:"balance"`
}

type GetWalletBalanceRequest struct {
	ID uuid.UUID `form:"id" json:"id"`
}

func ToWalletResponse(wallet models.Wallet) WalletResponse {
	return WalletResponse{
		ID:        wallet.ID,
		Address:   wallet.Address,
		Chain:     wallet.Chain,
		Label:     wallet.Label,
		UserID:    wallet.UserID,
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}
}

func ToWalletResponses(wallets []models.Wallet) []WalletResponse {
	resp := make([]WalletResponse, len(wallets))
	for i, wallet := range wallets {
		resp[i] = ToWalletResponse(wallet)
	}
	return resp
}

func ToGetWalletBalanceResponse(balance *core.Balance) GetWalletBalanceResponse {
	return GetWalletBalanceResponse{
		Chain:   balance.Chain,
		Address: balance.Address,
		Balance: balance.Balance,
	}
}
