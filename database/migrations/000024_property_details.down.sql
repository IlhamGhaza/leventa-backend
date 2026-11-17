ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_dist_nonneg;
ALTER TABLE properties DROP CONSTRAINT IF EXISTS chk_properties_furnishing;
ALTER TABLE properties
  DROP COLUMN IF EXISTS short_description, DROP COLUMN IF EXISTS area_label, DROP COLUMN IF EXISTS sub_district,
  DROP COLUMN IF EXISTS postal_code, DROP COLUMN IF EXISTS orientation, DROP COLUMN IF EXISTS maintenance_fee,
  DROP COLUMN IF EXISTS cert_shm, DROP COLUMN IF EXISTS cert_hgb, DROP COLUMN IF EXISTS cert_imb,
  DROP COLUMN IF EXISTS cert_pbb, DROP COLUMN IF EXISTS cert_girik, DROP COLUMN IF EXISTS cert_other,
  DROP COLUMN IF EXISTS living_rooms, DROP COLUMN IF EXISTS dining_rooms, DROP COLUMN IF EXISTS master_bedrooms,
  DROP COLUMN IF EXISTS child_bedrooms, DROP COLUMN IF EXISTS other_bedrooms, DROP COLUMN IF EXISTS master_bathrooms,
  DROP COLUMN IF EXISTS child_bathrooms, DROP COLUMN IF EXISTS other_bathrooms, DROP COLUMN IF EXISTS family_rooms,
  DROP COLUMN IF EXISTS study_rooms, DROP COLUMN IF EXISTS kitchens, DROP COLUMN IF EXISTS electricity_capacity,
  DROP COLUMN IF EXISTS air_conditioners, DROP COLUMN IF EXISTS tv_provider, DROP COLUMN IF EXISTS water_source,
  DROP COLUMN IF EXISTS furnishing_status, DROP COLUMN IF EXISTS has_pool, DROP COLUMN IF EXISTS has_garage,
  DROP COLUMN IF EXISTS has_front_yard, DROP COLUMN IF EXISTS has_back_yard, DROP COLUMN IF EXISTS has_balcony,
  DROP COLUMN IF EXISTS dist_school, DROP COLUMN IF EXISTS dist_bus_terminal, DROP COLUMN IF EXISTS dist_mrt_lrt,
  DROP COLUMN IF EXISTS dist_toll, DROP COLUMN IF EXISTS dist_hospital;
