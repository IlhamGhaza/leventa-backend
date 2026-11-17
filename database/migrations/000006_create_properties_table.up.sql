CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES owners(id),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    price DECIMAL(15,2),
    rental_price DECIMAL(15,2),
    description TEXT,
    address TEXT,
    province_id UUID REFERENCES provinces(id),
    city_id UUID REFERENCES cities(id),
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    land_area DECIMAL(10,2),
    building_area DECIMAL(10,2),
    floor INT,
    bedrooms INT DEFAULT 0,
    bathrooms INT DEFAULT 0,
    year_built INT,
    thumbnail VARCHAR(500),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_properties_code ON properties(code);
CREATE INDEX idx_properties_status ON properties(status);
CREATE INDEX idx_properties_type ON properties(type);
CREATE INDEX idx_properties_city_id ON properties(city_id);
CREATE INDEX idx_properties_province_id ON properties(province_id);
CREATE INDEX idx_properties_price ON properties(price);
CREATE INDEX idx_properties_owner_id ON properties(owner_id);
CREATE INDEX idx_properties_deleted_at ON properties(deleted_at);
