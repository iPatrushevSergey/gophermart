-- +goose Up
-- status: 0=NEW, 1=PROCESSING, 2=INVALID, 3=PROCESSED (mapped in adapter)
CREATE TABLE IF NOT EXISTS orders (
    id           BIGSERIAL PRIMARY KEY,
    number       TEXT NOT NULL UNIQUE,
    user_id      BIGINT NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status       SMALLINT NOT NULL DEFAULT 0,
    accrual      DOUBLE PRECISION,
    uploaded_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX idx_orders_user_id ON orders (user_id);
CREATE INDEX idx_orders_status_uploaded ON orders (status, uploaded_at);

-- +goose Down
DROP INDEX IF EXISTS idx_orders_status_uploaded;
DROP INDEX IF EXISTS idx_orders_user_id;
DROP TABLE IF EXISTS orders;
