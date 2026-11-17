CREATE TABLE IF NOT EXISTS units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_number VARCHAR(50) NOT NULL,
    floor INT,
    area DECIMAL(10,2),
    bedrooms INT DEFAULT 0,
    bathrooms INT DEFAULT 0,
    price DECIMAL(15,2),
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_status ON units(status);
