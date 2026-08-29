package domain

import (
	walletv1 "tracker/gen/wallet/v1"
)

type Wallet struct {
	ID      string `json:"id"`
	Address string `json:"address" `
	Chain   string `json:"chain" `
	Label   string `json:"label"`
	Symbol  string `json:"symbol"`
	UserID  string `json:"user_id"`
}

type WalletBalance struct {
	Wallet
	Balance    float64
	BalanceUSD float64
	HasError   bool
	ErrorMsg   string
	Price      TokenPrice
}

type WalletCreatedEvent struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
}

func (w *WalletBalance) ToGrpc() *walletv1.Wallet {
	return &walletv1.Wallet{
		Id:            w.Wallet.ID,
		Address:       w.Wallet.Address,
		Chain:         w.Wallet.Chain,
		TokenSymbol:   w.Wallet.Symbol,
		Label:         w.Wallet.Label,
		BalanceCrypto: float32(w.Balance),
		BalanceUsd:    float32(w.BalanceUSD),
		HasError:      w.HasError,
		ErrorMsg:      w.ErrorMsg,
		Price:         w.Price.ToGrpc(),
	}
}
