-- +goose Up
CREATE TABLE IF NOT EXISTS withdrawals (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    order_number TEXT NOT NULL,
    amount       DOUBLE PRECISION NOT NULL CHECK (amount > 0),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user_id ON withdrawals (user_id);
CREATE INDEX idx_withdrawals_processed_at ON withdrawals (processed_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_withdrawals_processed_at;
DROP INDEX IF EXISTS idx_withdrawals_user_id;
DROP TABLE IF EXISTS withdrawals;
