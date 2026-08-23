package domain

type NotificationCommand struct {
	UserID     string  `json:"user_id"`
	Email      string  `json:"email"`
	UserName   string  `json:"user_name"`
	AlertID    string  `json:"alert_id"`
	CoinName   string  `json:"coin_name"`
	CoinSymbol string  `json:"coin_symbol"`
	Price      float64 `json:"price"`
}
