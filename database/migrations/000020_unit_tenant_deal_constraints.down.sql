DROP INDEX IF EXISTS uq_commissions_deal_id;
DROP INDEX IF EXISTS uq_deals_lead_id;
DROP INDEX IF EXISTS idx_tenants_unit_id;
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS chk_tenants_status;
ALTER TABLE tenants DROP CONSTRAINT IF EXISTS chk_tenants_period;
ALTER TABLE units DROP CONSTRAINT IF EXISTS chk_units_status;
ALTER TABLE units DROP CONSTRAINT IF EXISTS uq_units_property_unit_number;
