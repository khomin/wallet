package cmd

type KnownWallets struct {
	UserID string       `json:"userID"`
	Tokens []KnownToken `json:"tokens"`
}

type KnownToken struct {
	Chain  string `json:"chains"`
	Symbol string `json:"symbol"`
	Addr   string `json:"addr"`
	Label  string `json:"label"`
}
