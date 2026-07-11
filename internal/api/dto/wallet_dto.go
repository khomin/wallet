package dto

import (
	"time"

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
	Address string `json:"address" binding:"required"`
	Chain   string `json:"chain" binding:"required"`
	Label   string `json:"label,omitempty"`
	UserID  string `json:"user_id,omitempty"`
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
