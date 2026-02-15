CREATE TABLE IF NOT EXISTS balance_accounts (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    current         DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (current >= 0),
    withdrawn_total DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (withdrawn_total >= 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version         BIGINT NOT NULL DEFAULT 0
);
