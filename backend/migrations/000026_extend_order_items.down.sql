ALTER TABLE order_items DROP CONSTRAINT IF EXISTS chk_order_items_product_or_name;
ALTER TABLE order_items DROP COLUMN IF EXISTS custom_item_description;
ALTER TABLE order_items DROP COLUMN IF EXISTS custom_item_name;
ALTER TABLE order_items ALTER COLUMN product_id SET NOT NULL;
