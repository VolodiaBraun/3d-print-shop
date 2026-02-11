CREATE TABLE delivery_zones (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    city VARCHAR(255) NOT NULL,
    region VARCHAR(255),
    delivery_cost DECIMAL(10,2) NOT NULL DEFAULT 0,
    free_delivery_from DECIMAL(10,2),
    estimated_days_min INTEGER NOT NULL DEFAULT 1,
    estimated_days_max INTEGER NOT NULL DEFAULT 3,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_delivery_zones_city ON delivery_zones(city);
CREATE INDEX idx_delivery_zones_active ON delivery_zones(is_active);
