DROP INDEX IF EXISTS idx_orders_payment_provider_id;
DROP INDEX IF EXISTS idx_orders_order_type;

ALTER TABLE orders
    DROP COLUMN IF EXISTS order_type,
    DROP COLUMN IF EXISTS payment_link,
    DROP COLUMN IF EXISTS payment_provider,
    DROP COLUMN IF EXISTS payment_provider_id,
    DROP COLUMN IF EXISTS payment_expires_at;
