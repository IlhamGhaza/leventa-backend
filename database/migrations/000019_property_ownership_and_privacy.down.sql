DROP TABLE IF EXISTS property_marketing_assignments;

ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_status;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_transaction_type;
DROP INDEX IF EXISTS idx_properties_land_area;
DROP INDEX IF EXISTS idx_properties_building_area;
DROP INDEX IF EXISTS idx_properties_transaction_type;
DROP INDEX IF EXISTS idx_properties_primary_marketing_id;

ALTER TABLE properties ADD COLUMN IF NOT EXISTS address TEXT;
UPDATE properties SET address = street_name;
ALTER TABLE properties DROP COLUMN IF EXISTS room_number;
ALTER TABLE properties DROP COLUMN IF EXISTS floor_number;
ALTER TABLE properties DROP COLUMN IF EXISTS house_number;
ALTER TABLE properties DROP COLUMN IF EXISTS street_name;
ALTER TABLE properties DROP COLUMN IF EXISTS cluster_name;
ALTER TABLE properties DROP COLUMN IF EXISTS block_number;
ALTER TABLE properties DROP COLUMN IF EXISTS general_location;
ALTER TABLE properties DROP COLUMN IF EXISTS transaction_type;
ALTER TABLE properties DROP COLUMN IF EXISTS primary_marketing_id;

DROP INDEX IF EXISTS idx_owners_created_by;
ALTER TABLE owners DROP COLUMN IF EXISTS created_by;
ALTER TABLE owners RENAME COLUMN private_address TO address;

ALTER TABLE users DROP COLUMN IF EXISTS whatsapp_number;
