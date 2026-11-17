-- Property ownership, private address split, owner privacy, WhatsApp contact.
-- NOTE: existing properties (if any) are backfilled with the oldest active
-- `sales` user as primary marketing. The migration fails loudly when properties
-- exist but no sales user exists, because ownership cannot be inferred.

ALTER TABLE users ADD COLUMN IF NOT EXISTS whatsapp_number VARCHAR(20);

-- Owners: private contact data, scoped to the marketing user that created it.
ALTER TABLE owners RENAME COLUMN address TO private_address;
ALTER TABLE owners ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users(id);
CREATE INDEX IF NOT EXISTS idx_owners_created_by ON owners(created_by);

-- Properties
ALTER TABLE properties ADD COLUMN IF NOT EXISTS primary_marketing_id UUID REFERENCES users(id);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS transaction_type VARCHAR(10) NOT NULL DEFAULT 'sale';
ALTER TABLE properties ADD COLUMN IF NOT EXISTS general_location VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS block_number VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cluster_name VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS street_name VARCHAR(255);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS house_number VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS floor_number VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS room_number VARCHAR(50);

-- Legacy single address column is restricted data: move it to street_name.
UPDATE properties SET street_name = LEFT(address, 255) WHERE address IS NOT NULL AND street_name IS NULL;
ALTER TABLE properties DROP COLUMN IF EXISTS address;

UPDATE properties SET transaction_type = 'rent' WHERE price IS NULL AND rental_price IS NOT NULL;

DO $$
DECLARE fallback_id UUID;
BEGIN
    IF EXISTS (SELECT 1 FROM properties WHERE primary_marketing_id IS NULL) THEN
        SELECT id INTO fallback_id FROM users
         WHERE role = 'sales' AND status = 'active' AND deleted_at IS NULL
         ORDER BY created_at LIMIT 1;
        IF fallback_id IS NULL THEN
            RAISE EXCEPTION 'cannot backfill properties.primary_marketing_id: no active sales user exists';
        END IF;
        UPDATE properties SET primary_marketing_id = fallback_id WHERE primary_marketing_id IS NULL;
        UPDATE owners o SET created_by = fallback_id WHERE o.created_by IS NULL;
    END IF;
END $$;

ALTER TABLE properties ALTER COLUMN primary_marketing_id SET NOT NULL;
ALTER TABLE properties ADD CONSTRAINT chk_properties_transaction_type CHECK (transaction_type IN ('sale', 'rent'));
ALTER TABLE properties ADD CONSTRAINT chk_properties_status
    CHECK (status IN ('draft', 'available', 'reserved', 'sold', 'rented', 'maintenance', 'archived'));

CREATE INDEX IF NOT EXISTS idx_properties_primary_marketing_id ON properties(primary_marketing_id);
CREATE INDEX IF NOT EXISTS idx_properties_transaction_type ON properties(transaction_type);
CREATE INDEX IF NOT EXISTS idx_properties_land_area ON properties(land_area);
CREATE INDEX IF NOT EXISTS idx_properties_building_area ON properties(building_area);

-- Marketing assignments (1 primary + max 2 co-listing, enforced in a locked transaction).
CREATE TABLE IF NOT EXISTS property_marketing_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    marketing_id UUID NOT NULL REFERENCES users(id),
    assignment_type VARCHAR(20) NOT NULL CHECK (assignment_type IN ('primary', 'co_listing')),
    assigned_by UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_pma_property_marketing UNIQUE (property_id, marketing_id)
);

CREATE UNIQUE INDEX uq_pma_one_primary_per_property
    ON property_marketing_assignments(property_id) WHERE assignment_type = 'primary';
CREATE INDEX idx_pma_property_id ON property_marketing_assignments(property_id);
CREATE INDEX idx_pma_marketing_id ON property_marketing_assignments(marketing_id);

INSERT INTO property_marketing_assignments (property_id, marketing_id, assignment_type, assigned_by)
SELECT id, primary_marketing_id, 'primary', primary_marketing_id FROM properties;
