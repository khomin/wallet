-- Create a lightweight coin catalog table for metadata and identity.
CREATE TABLE IF NOT EXISTS coins (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coin_id TEXT NOT NULL UNIQUE,
    symbol TEXT NOT NULL,
    coin_name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    snapshot_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create a dedicated price snapshot table for price/time-series data.
CREATE TABLE IF NOT EXISTS coin_price_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coin_id TEXT NOT NULL UNIQUE,
    symbol TEXT NOT NULL,
    coin_name TEXT NOT NULL,
    price_usd DECIMAL(40,18) NOT NULL,
    market_cap_usd DECIMAL(40,18) NOT NULL,
    total_volume_usd DECIMAL(40,18) NOT NULL,
    price_change_24h DECIMAL(40,18) NOT NULL,
    price_change_percent_24h DECIMAL(16,4) NOT NULL,
    market_cap_change_24h DECIMAL(40,18) NOT NULL,
    market_cap_change_percent_24h DECIMAL(16,4) NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    snapshot_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_coins_coin_id ON coins(coin_id);
CREATE INDEX idx_coins_symbol ON coins(symbol);
CREATE INDEX idx_coins_snapshot_at ON coins(snapshot_at);

CREATE INDEX idx_coin_price_snapshots_coin_id ON coin_price_snapshots(coin_id);
CREATE INDEX idx_coin_price_snapshots_symbol ON coin_price_snapshots(symbol);
CREATE INDEX idx_coin_price_snapshots_snapshot_at ON coin_price_snapshots(snapshot_at);
CREATE INDEX idx_coin_price_snapshots_price_usd ON coin_price_snapshots(price_usd);