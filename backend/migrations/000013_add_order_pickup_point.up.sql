ALTER TABLE orders ADD COLUMN pickup_point_id INTEGER REFERENCES pickup_points(id);
ALTER TABLE orders ADD COLUMN delivery_provider VARCHAR(50);
ALTER TABLE orders ADD COLUMN estimated_delivery VARCHAR(50);
