ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS order_type          VARCHAR(20)              NOT NULL DEFAULT 'regular',
    ADD COLUMN IF NOT EXISTS payment_link        VARCHAR(500),
    ADD COLUMN IF NOT EXISTS payment_provider    VARCHAR(50),
    ADD COLUMN IF NOT EXISTS payment_provider_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS payment_expires_at  TIMESTAMP WITH TIME ZONE;

CREATE INDEX IF NOT EXISTS idx_orders_order_type
    ON orders (order_type);

CREATE INDEX IF NOT EXISTS idx_orders_payment_provider_id
    ON orders (payment_provider_id)
    WHERE payment_provider_id IS NOT NULL;
