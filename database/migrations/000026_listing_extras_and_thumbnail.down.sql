DROP INDEX IF EXISTS uq_property_images_one_thumbnail;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_extra_distances_array;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_extra_outdoor_array;
ALTER TABLE properties DROP COLUMN IF EXISTS extra_distances, DROP COLUMN IF EXISTS extra_outdoor;
