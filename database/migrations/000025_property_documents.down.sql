DROP TABLE IF EXISTS property_documents;
ALTER TABLE properties DROP COLUMN IF EXISTS cert_slf, DROP COLUMN IF EXISTS cert_hgu, DROP COLUMN IF EXISTS cert_shmsrs;
