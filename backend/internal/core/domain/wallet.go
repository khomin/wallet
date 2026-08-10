package domain

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
