CREATE TABLE custom_order_details (
    id                 SERIAL PRIMARY KEY,
    order_id           INTEGER UNIQUE NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    client_description TEXT,
    admin_notes        TEXT,
    file_urls          JSONB NOT NULL DEFAULT '[]',
    print_settings     JSONB NOT NULL DEFAULT '{}',
    bitrix_deal_id     VARCHAR(100),
    bitrix_stage_id    VARCHAR(100),
    created_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_custom_order_details_order_id
    ON custom_order_details (order_id);

CREATE INDEX idx_custom_order_details_bitrix_deal_id
    ON custom_order_details (bitrix_deal_id)
    WHERE bitrix_deal_id IS NOT NULL;
