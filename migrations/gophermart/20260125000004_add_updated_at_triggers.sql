-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TRIGGER set_balance_accounts_updated_at
    BEFORE UPDATE ON balance_accounts
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TRIGGER set_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS set_orders_updated_at ON orders;
DROP TRIGGER IF EXISTS set_balance_accounts_updated_at ON balance_accounts;
DROP TRIGGER IF EXISTS set_users_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();
