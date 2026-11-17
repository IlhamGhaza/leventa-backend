CREATE TABLE IF NOT EXISTS cities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID NOT NULL REFERENCES provinces(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(province_id, name)
);

CREATE INDEX idx_cities_province_id ON cities(province_id);
