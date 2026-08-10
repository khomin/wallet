CREATE TABLE IF NOT EXISTS coins (
    id TEXT PRIMARY KEY,
    symbol TEXT NOT NULL,
    coin_name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS coin_prices (
    id TEXT PRIMARY KEY REFERENCES coins(id),
    price_usd DECIMAL(40,18) NOT NULL,
    market_cap_usd DECIMAL(40,18) NOT NULL,
    total_volume_usd DECIMAL(40,18) NOT NULL,
    price_change_24h DECIMAL(40,18) NOT NULL,
    price_change_percent_24h DECIMAL(16,4) NOT NULL,
    market_cap_change_24h DECIMAL(40,18) NOT NULL,
    market_cap_change_percent_24h DECIMAL(16,4) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_coins_symbol ON coins(symbol);

CREATE INDEX idx_coin_prices_price_usd ON coin_prices(price_usd);