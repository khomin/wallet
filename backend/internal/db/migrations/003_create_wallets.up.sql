CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
    address TEXT NOT NULL,
    chain TEXT NOT NULL,
    coin_id TEXT NOT NULL REFERENCES coins(id) ON DELETE CASCADE,
    label TEXT,
    notify BOOLEAN DEFAULT TRUE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_balances (
    id UUID PRIMARY KEY REFERENCES wallets(id) ON DELETE CASCADE,
    value_crypto DECIMAL(40,18) NOT NULL,
    value_usd DECIMAL(40,18) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_balance_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID REFERENCES wallets(id) ON DELETE CASCADE,
    value_crypto DECIMAL(40,18) NOT NULL,
    value_usd DECIMAL(40,18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_address_chain 
    ON wallets (address, chain, coin_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_updated 
    ON wallets (updated_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_balance_updated_at
    ON wallet_balances (updated_at);

CREATE UNIQUE INDEX IF NOT EXISTS idx_wallet_balance_snapshots_created_at
    ON wallet_balance_snapshots (created_at);