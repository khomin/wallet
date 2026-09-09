package domain

import (
	"time"
	walletv1 "tracker/gen/wallet/v1"
)

type Wallet struct {
	ID      string `json:"id"`
	Address string `json:"address" `
	Chain   string `json:"chain" `
	Symbol  string `json:"symbol"`
}

type WalletBalance struct {
	Wallet
	Balance    float64
	BalanceUSD float64
	UpdatedAt  time.Time
	HasError   bool
	ErrorMsg   string
	Price      TokenPrice
}

type UserWallet struct {
	Wallet
	UserID string `json:"user_id"`
	Label  string `json:"label"`
	Notify bool   `json:"notify"`
}

type UserWalletBalance struct {
	UserWallet
	Balance    float64
	BalanceUSD float64
	UpdatedAt  time.Time
	HasError   bool
	ErrorMsg   string
}

type WalletBalanceSnapshot struct {
	Balance    float64
	BalanceUSD float64
	Time       time.Time
}

type WalletBalanceChange struct {
	Wallet
	CurrentBalance    float64
	CurrentBalanceUSD float64
	OldBalance        float64
	OldBalanceUSD     float64
}

func (w *UserWalletBalance) ToGrpc() *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:            w.UserWallet.ID,
		Address:       w.UserWallet.Address,
		Chain:         w.UserWallet.Chain,
		TokenSymbol:   w.UserWallet.Symbol,
		Label:         w.UserWallet.Label,
		Notify:        w.UserWallet.Notify,
		BalanceCrypto: w.Balance,
		BalanceUsd:    w.BalanceUSD,
		HasError:      w.HasError,
		ErrorMsg:      w.ErrorMsg,
	}
}
