DROP TABLE IF EXISTS property_facilities;
DROP TABLE IF EXISTS facilities;
DROP INDEX IF EXISTS idx_properties_district_id;
ALTER TABLE properties DROP COLUMN IF EXISTS district_id;
DROP TABLE IF EXISTS districts;
