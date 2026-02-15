CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TRIGGER set_balance_accounts_updated_at
    BEFORE UPDATE ON balance_accounts
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();

CREATE TRIGGER set_orders_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE PROCEDURE set_updated_at();
