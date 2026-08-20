package domain

import (
	pricev1 "tracker/gen/price/v1"
	walletv1 "tracker/gen/wallet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type Wallet struct {
	ID      string `json:"id"`
	Address string `json:"address" `
	Chain   string `json:"chain" `
	Label   string `json:"label"`
	Symbol  string `json:"symbol"`
	UserID  string `json:"user_id"`
}

type WalletWithBalance struct {
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

func (w *WalletWithBalance) ToGrpc() *walletv1.Wallet {
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
		Price: &pricev1.Price{
			Symbol:                        w.Symbol,
			Name:                          w.Price.Name,
			PriceUsd:                      float32(w.Price.CurrentPrice),
			MarketCap:                     float32(w.Price.MarketCap),
			TotalVolume:                   float32(w.Price.TotalVolume),
			High_24H:                      float32(w.Price.High_24h),
			Low_24H:                       float32(w.Price.Low_24h),
			PriceChange_24H:               float32(w.Price.PriceChange_24h),
			PriceChangePercentage_24H:     float32(w.Price.PriceChangePercentage_24h),
			MarketCapChange_24H:           float32(w.Price.MarketCapChange_24h),
			MarketCapChangePercentage_24H: float32(w.Price.MarketCapChange_percentage_24h),
			UpdatedAt:                     timestamppb.New(w.Price.UpdatedAt),
		},
	}
}
