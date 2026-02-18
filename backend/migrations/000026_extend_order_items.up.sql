-- Make product_id nullable to support custom order line items (no catalog product).
ALTER TABLE order_items ALTER COLUMN product_id DROP NOT NULL;

-- Free-text item name and description for custom orders.
ALTER TABLE order_items
    ADD COLUMN IF NOT EXISTS custom_item_name        VARCHAR(255),
    ADD COLUMN IF NOT EXISTS custom_item_description TEXT;

-- Enforce: either product_id or custom_item_name must be present.
ALTER TABLE order_items
    ADD CONSTRAINT chk_order_items_product_or_name
    CHECK (product_id IS NOT NULL OR custom_item_name IS NOT NULL);
