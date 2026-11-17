-- Custom outdoor areas and nearby places entered by the primary marketing,
-- next to the fixed has_* / dist_* columns (which stay for filtering).
ALTER TABLE properties ADD COLUMN IF NOT EXISTS extra_outdoor JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS extra_distances JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE properties ADD CONSTRAINT chk_properties_extra_outdoor_array CHECK (jsonb_typeof(extra_outdoor) = 'array');
ALTER TABLE properties ADD CONSTRAINT chk_properties_extra_distances_array CHECK (jsonb_typeof(extra_distances) = 'array');

-- Thumbnail: at most one per property, mirrored into properties.thumbnail so
-- list endpoints (which do not load images) can show a cover.
UPDATE property_images SET is_thumbnail = FALSE
WHERE is_thumbnail AND id NOT IN (
    SELECT DISTINCT ON (property_id) id FROM property_images
    WHERE is_thumbnail ORDER BY property_id, sort_order, created_at, id);

UPDATE property_images SET is_thumbnail = TRUE
WHERE id IN (
    SELECT DISTINCT ON (property_id) id FROM property_images pi
    WHERE NOT EXISTS (SELECT 1 FROM property_images t WHERE t.property_id = pi.property_id AND t.is_thumbnail)
    ORDER BY property_id, sort_order, created_at, id);

UPDATE properties p SET thumbnail = i.url
FROM property_images i WHERE i.property_id = p.id AND i.is_thumbnail;

CREATE UNIQUE INDEX IF NOT EXISTS uq_property_images_one_thumbnail
    ON property_images (property_id) WHERE is_thumbnail;
