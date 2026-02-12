CREATE TABLE sales_daily_stats (
    id             SERIAL PRIMARY KEY,
    date           DATE NOT NULL UNIQUE,
    revenue        DECIMAL(12,2) NOT NULL DEFAULT 0,
    orders_count   INTEGER NOT NULL DEFAULT 0,
    avg_check      DECIMAL(10,2) NOT NULL DEFAULT 0,
    new_customers  INTEGER NOT NULL DEFAULT 0,
    items_sold     INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sales_daily_stats_date ON sales_daily_stats(date);
