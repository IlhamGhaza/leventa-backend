ALTER TABLE units ADD CONSTRAINT uq_units_property_unit_number UNIQUE (property_id, unit_number);
ALTER TABLE units ADD CONSTRAINT chk_units_status CHECK (status IN ('available', 'reserved', 'sold', 'rented', 'maintenance'));

ALTER TABLE tenants ADD CONSTRAINT chk_tenants_period CHECK (end_date IS NULL OR end_date >= start_date);
ALTER TABLE tenants ADD CONSTRAINT chk_tenants_status CHECK (status IN ('active', 'expired', 'terminated'));
CREATE INDEX IF NOT EXISTS idx_tenants_unit_id ON tenants(unit_id);

-- A lead yields at most one deal, a deal at most one commission.
CREATE UNIQUE INDEX IF NOT EXISTS uq_deals_lead_id ON deals(lead_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_commissions_deal_id ON commissions(deal_id);
