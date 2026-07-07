-- Create the coin snapshots table
CREATE TABLE IF NOT EXISTS coin_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    coin_id TEXT NOT NULL UNIQUE,
    symbol TEXT NOT NULL,
    coin_name TEXT NOT NULL,
    
    -- Market Data
    price_usd DECIMAL(40,18) NOT NULL,
    market_cap_usd DECIMAL(40,18) NOT NULL,
    market_cap_rank INTEGER NOT NULL,
    total_volume_usd DECIMAL(40,18) NOT NULL,
    
    -- 24h Changes
    price_change_24h DECIMAL(40,18) NOT NULL,
    price_change_percent_24h DECIMAL(16,4) NOT NULL,
    market_cap_change_24h DECIMAL(40,18) NOT NULL,
    market_cap_change_percent_24h DECIMAL(16,4) NOT NULL,
    
    -- Supply
    circulating_supply DECIMAL(40,18) NOT NULL,
    total_supply DECIMAL(40,18),
    max_supply DECIMAL(40,18),
    
    -- All-Time High/Low
    ath DECIMAL(40,18) NOT NULL,
    ath_change_percent DECIMAL(16,4) NOT NULL,
    ath_date TIMESTAMP NOT NULL,
    atl DECIMAL(40,18) NOT NULL,
    atl_change_percent DECIMAL(16,4) NOT NULL,
    atl_date TIMESTAMP NOT NULL,
    
    -- Metadata
    image_url TEXT NOT NULL,
    last_updated TIMESTAMP NOT NULL,
    snapshot_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for fast queries
CREATE INDEX idx_coin_snapshots_coin_id ON coin_snapshots(coin_id);
CREATE INDEX idx_coin_snapshots_symbol ON coin_snapshots(symbol);
CREATE INDEX idx_coin_snapshots_snapshot_at ON coin_snapshots(snapshot_at);
CREATE INDEX idx_coin_snapshots_price_usd ON coin_snapshots(price_usd);

-- Composite index for getting latest snapshots
CREATE INDEX idx_coin_snapshots_coin_snapshot ON coin_snapshots(coin_id, snapshot_at DESC);