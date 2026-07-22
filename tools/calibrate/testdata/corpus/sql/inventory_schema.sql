CREATE TABLE warehouse.location (
    location_id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    code TEXT NOT NULL UNIQUE CHECK (code ~ '^[A-Z]{2}-[0-9]{3}$'),
    description TEXT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE warehouse.stock_item (
    sku TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    unit_cost NUMERIC(12, 2) NOT NULL CHECK (unit_cost >= 0),
    reorder_level INTEGER NOT NULL DEFAULT 0 CHECK (reorder_level >= 0)
);

CREATE TABLE warehouse.stock_balance (
    location_id BIGINT NOT NULL REFERENCES warehouse.location(location_id),
    sku TEXT NOT NULL REFERENCES warehouse.stock_item(sku),
    quantity INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (location_id, sku)
);

CREATE INDEX stock_balance_reorder_idx
    ON warehouse.stock_balance (sku, quantity)
    WHERE quantity >= 0;

COMMENT ON TABLE warehouse.stock_balance IS
    'Current on-hand quantity for each SKU and warehouse location';
