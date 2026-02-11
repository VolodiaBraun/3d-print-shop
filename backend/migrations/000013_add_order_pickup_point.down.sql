ALTER TABLE orders DROP COLUMN IF EXISTS pickup_point_id;
ALTER TABLE orders DROP COLUMN IF EXISTS delivery_provider;
ALTER TABLE orders DROP COLUMN IF EXISTS estimated_delivery;
