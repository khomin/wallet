package demo

import (
	"tracker/internal/core/domain"

	"github.com/google/uuid"
)

type DemoWallets struct {
	Wallets map[string]domain.WalletBalance
}

func NewDemoWallets() *DemoWallets {
	var wallets = make(map[string]domain.WalletBalance)
	for _, i := range walletList {
		wallets[i.ID] = i
	}
	return &DemoWallets{
		Wallets: wallets,
	}
}

func (d *DemoWallets) GetWallets() []domain.WalletBalance {
	return walletList
}

func (d *DemoWallets) GetWallet(id uuid.UUID) (*domain.WalletBalance, error) {
	v, found := d.Wallets[id.String()]
	if !found {
		return nil, domain.ErrorNotFound
	}
	return &v, nil
}

var walletList = []domain.WalletBalance{
	{
		Wallet: domain.Wallet{
			ID:      "4e4e27d5-b47c-4584-bac9-a998e3d87980",
			Address: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", // Vitalik's address
			Chain:   "ETH",
			Symbol:  "ETH",
			Label:   "Main Treasury (Demo)",
		},
		Balance:    124.55,
		BalanceUSD: 398560.20,
		HasError:   false,
	},
	{
		Wallet: domain.Wallet{
			ID:      "4e4e27d5-b47c-4584-bac9-a998e3d87981",
			Address: "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa", // Satoshi's genesis address
			Chain:   "BTC",
			Symbol:  "BTC",
			Label:   "Cold Storage (Demo)",
		},
		Balance:    12.4,
		BalanceUSD: 793600.00,
		HasError:   false,
	},
	{
		Wallet: domain.Wallet{
			ID:      "4e4e27d5-b47c-4584-bac9-a998e3d87982",
			Address: "7xKXtg2CW87d97TXJSDpbD5jBkheTqA83TZRuJosgAsU",
			Chain:   "SOL",
			Symbol:  "SOL",
			Label:   "DeFi Staking (Demo)",
		},
		Balance:    450.00,
		BalanceUSD: 67500.00,
		HasError:   false,
	},
}
