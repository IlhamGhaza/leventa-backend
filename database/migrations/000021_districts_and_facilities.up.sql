CREATE TABLE IF NOT EXISTS districts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    city_id UUID NOT NULL REFERENCES cities(id),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_districts_city_name UNIQUE (city_id, name)
);
CREATE INDEX idx_districts_city_id ON districts(city_id);

ALTER TABLE properties ADD COLUMN IF NOT EXISTS district_id UUID REFERENCES districts(id);
CREATE INDEX IF NOT EXISTS idx_properties_district_id ON properties(district_id);

CREATE TABLE IF NOT EXISTS facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS property_facilities (
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    facility_id UUID NOT NULL REFERENCES facilities(id),
    PRIMARY KEY (property_id, facility_id)
);
CREATE INDEX idx_property_facilities_facility_id ON property_facilities(facility_id);
