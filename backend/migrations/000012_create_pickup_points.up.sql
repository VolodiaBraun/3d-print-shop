CREATE TABLE pickup_points (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(255) NOT NULL,
    latitude DECIMAL(10,7),
    longitude DECIMAL(10,7),
    phone VARCHAR(30),
    working_hours VARCHAR(255) NOT NULL DEFAULT 'Пн-Пт 10:00-20:00',
    provider VARCHAR(50) NOT NULL DEFAULT 'mock',
    external_id VARCHAR(255),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_pickup_points_city ON pickup_points(city);
CREATE INDEX idx_pickup_points_active ON pickup_points(is_active);
CREATE INDEX idx_pickup_points_provider ON pickup_points(provider);
