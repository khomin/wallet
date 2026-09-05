package demo

import (
	"time"
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

func (d *DemoWallets) GetWalletBalanceSnapshot(
	id uuid.UUID,
	from time.Time,
	to time.Time,
) ([]domain.WalletBalanceSnapshot, error) {
	_, found := d.Wallets[id.String()]
	if !found {
		return nil, domain.ErrorNotFound
	}
	now := time.Now()
	start := now.AddDate(-5, 0, 0)
	balanceCrypto := 123.0
	balanceUSD := 456.0
	// Deterministic changes: realistic-ish movement without random tests.
	changes := []float64{
		+80, -12, +3, -70, +150, -4, +2, +1, -30, +90,
		-110, +7, -2, +45, -8, +200, -150, +4, +3, -2,
		+15, -5, -90, +180, -20, +2, -3, +5, +120, -200,
		+8, +6, -4, -5, +250, -30, -180, +12, +3, +170,
		-60, +2, -2, +4, -300, +50, +8, +9, -20, +220,
		-100, +5, -3, +2, +400, -350, +20, -10, +5, +3,
		-40, +180, -170, +15, +2, -3, +90, -20, +250, -200,
		+4, +5, -7, +600, -500, +10, -3, +2, +30, -25,
		+200, -150, +8, -5, +3, -300, +450, -50, +2, -1,
		+700, -650, +20, -10, +5, +2, -100, +250, -200, +80,
	}
	out := make([]domain.WalletBalanceSnapshot, 0)
	i := 0
	for t := start; !t.After(now); t = t.Add(6 * time.Hour) {
		change := changes[i%len(changes)]
		balanceCrypto += change
		// Some independent USD movement so the two charts aren't identical.
		balanceUSD += change*10 + float64((i%7)-3)*8
		if !t.Before(from) && !t.After(to) {
			out = append(out, domain.WalletBalanceSnapshot{
				Balance:    balanceCrypto,
				BalanceUSD: balanceUSD,
				Time:       t,
			})
		}
		i++
	}
	return out, nil
}
