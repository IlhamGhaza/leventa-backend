-- D-028: listing fields following the reference listing page.
ALTER TABLE properties ADD COLUMN IF NOT EXISTS short_description VARCHAR(500);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS area_label VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS sub_district VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS postal_code VARCHAR(10);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS orientation VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS maintenance_fee DECIMAL(15,2);

-- Ownership documents
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_shm BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_hgb BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_imb BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_pbb BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_girik BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_other BOOLEAN NOT NULL DEFAULT FALSE;

-- Room breakdown (bedrooms/bathrooms stay as totals used for filtering)
ALTER TABLE properties ADD COLUMN IF NOT EXISTS living_rooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dining_rooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS master_bedrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS child_bedrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS other_bedrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS master_bathrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS child_bathrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS other_bathrooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS family_rooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS study_rooms INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS kitchens INT NOT NULL DEFAULT 0;

-- Facilities
ALTER TABLE properties ADD COLUMN IF NOT EXISTS electricity_capacity VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS air_conditioners INT NOT NULL DEFAULT 0;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS tv_provider VARCHAR(100);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS water_source VARCHAR(100);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS furnishing_status VARCHAR(30);

-- Outdoor
ALTER TABLE properties ADD COLUMN IF NOT EXISTS has_pool BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS has_garage BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS has_front_yard BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS has_back_yard BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS has_balcony BOOLEAN NOT NULL DEFAULT FALSE;

-- Distance to public facilities, in meters
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dist_school INT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dist_bus_terminal INT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dist_mrt_lrt INT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dist_toll INT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS dist_hospital INT;

ALTER TABLE properties ADD CONSTRAINT chk_properties_furnishing
    CHECK (furnishing_status IS NULL OR furnishing_status IN ('', 'unfurnished', 'semi_furnished', 'furnished'));
ALTER TABLE properties ADD CONSTRAINT chk_properties_dist_nonneg CHECK (
    COALESCE(dist_school, 0) >= 0 AND COALESCE(dist_bus_terminal, 0) >= 0 AND COALESCE(dist_mrt_lrt, 0) >= 0
    AND COALESCE(dist_toll, 0) >= 0 AND COALESCE(dist_hospital, 0) >= 0);
