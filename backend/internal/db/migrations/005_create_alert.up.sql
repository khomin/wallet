CREATE TABLE IF NOT EXISTS alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    coin_id TEXT NOT NULL REFERENCES coins(id) ON DELETE CASCADE,
    condition TEXT NOT NULL CHECK (condition IN ('above', 'below')),
    price NUMERIC(30, 10) NOT NULL CHECK (price > 0),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    triggered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_user_id
    ON alerts(user_id);

CREATE INDEX IF NOT EXISTS idx_alerts_active
    ON alerts(enabled)
    WHERE enabled = TRUE;