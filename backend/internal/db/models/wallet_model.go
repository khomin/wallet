package models

import (
	"time"
	"tracker/internal/core/domain"

	"github.com/google/uuid"
)

type Wallet struct {
	ID        uuid.UUID
	Address   string
	Chain     string
	Label     string
	Symbol    string
	UserID    string
	UpdatedAt time.Time
}

type WalletBalance struct {
	Wallet
	Price      domain.TokenPrice
	Balance    float64
	BalanceUSD float64
	HasError   bool
	ErrorMsg   string
}
