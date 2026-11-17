-- D-026/D-027: internal system. No customers/staff, no leads/inquiries/
-- favorites/surveys/units/tenants/notifications. Admin records a Deal or
-- Cancelled outcome directly on a property.

-- 1. Deals hang on the property, not on a lead.
ALTER TABLE deals ADD COLUMN IF NOT EXISTS recorded_by UUID REFERENCES users(id);
ALTER TABLE deals ADD COLUMN IF NOT EXISTS outcome VARCHAR(20) NOT NULL DEFAULT 'deal';
ALTER TABLE deals ADD COLUMN IF NOT EXISTS buyer_name VARCHAR(255);

UPDATE deals d SET buyer_name = u.name
  FROM leads l JOIN users u ON u.id = l.customer_id
 WHERE d.lead_id = l.id AND d.buyer_name IS NULL;
UPDATE deals SET buyer_name = 'Unknown' WHERE buyer_name IS NULL;
UPDATE deals SET recorded_by = sales_id WHERE recorded_by IS NULL;

ALTER TABLE deals ALTER COLUMN recorded_by SET NOT NULL;
ALTER TABLE deals ALTER COLUMN type DROP NOT NULL;
ALTER TABLE deals ALTER COLUMN price DROP NOT NULL;
ALTER TABLE deals DROP COLUMN lead_id;

ALTER TABLE deals ADD CONSTRAINT chk_deals_outcome CHECK (outcome IN ('deal', 'cancelled'));
ALTER TABLE deals ADD CONSTRAINT chk_deals_deal_fields CHECK (
    outcome <> 'deal' OR (type IN ('sale', 'rent') AND price IS NOT NULL AND buyer_name IS NOT NULL)
);
-- One outcome per property.
CREATE UNIQUE INDEX IF NOT EXISTS uq_deals_property_id ON deals(property_id);
CREATE INDEX IF NOT EXISTS idx_deals_outcome ON deals(outcome);

-- 2. Drop out-of-scope tables.
DROP TABLE IF EXISTS tenants;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS lead_follow_ups;
DROP TABLE IF EXISTS surveys;
DROP TABLE IF EXISTS leads;
DROP TABLE IF EXISTS inquiries;
DROP TABLE IF EXISTS favorites;
DROP TABLE IF EXISTS notifications;

-- 3. Only Marketing (sales) and Admin accounts remain.
-- Customer and staff accounts are removed; their audit rows keep a NULL actor.
UPDATE audit_logs SET user_id = NULL
 WHERE user_id IN (SELECT id FROM users WHERE role IN ('customer', 'staff'));
DELETE FROM users WHERE role IN ('customer', 'staff');
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'sales';
ALTER TABLE users ADD CONSTRAINT chk_users_role CHECK (role IN ('sales', 'admin'));
