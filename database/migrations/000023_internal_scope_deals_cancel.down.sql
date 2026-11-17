-- Best-effort rollback: recreates the dropped tables EMPTY (data is not
-- restored) and the old deals shape. Customer/staff users are not restored.
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_users_role;
ALTER TABLE users ALTER COLUMN role SET DEFAULT 'customer';

CREATE TABLE IF NOT EXISTS units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    unit_number VARCHAR(50) NOT NULL,
    floor INT,
    area DECIMAL(10,2),
    bedrooms INT DEFAULT 0,
    bathrooms INT DEFAULT 0,
    price DECIMAL(15,2),
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_units_property_id ON units(property_id);
CREATE INDEX idx_units_status ON units(status);

CREATE TABLE IF NOT EXISTS favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, property_id)
);

CREATE INDEX idx_favorites_user_id ON favorites(user_id);
CREATE INDEX idx_favorites_property_id ON favorites(property_id);

CREATE TABLE IF NOT EXISTS inquiries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    message TEXT,
    preferred_contact VARCHAR(50),
    preferred_visit_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'new',
    assigned_sales_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inquiries_customer_id ON inquiries(customer_id);
CREATE INDEX idx_inquiries_property_id ON inquiries(property_id);
CREATE INDEX idx_inquiries_status ON inquiries(status);
CREATE INDEX idx_inquiries_assigned_sales_id ON inquiries(assigned_sales_id);

CREATE TABLE IF NOT EXISTS leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    inquiry_id UUID REFERENCES inquiries(id),
    customer_id UUID NOT NULL REFERENCES users(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    sales_id UUID REFERENCES users(id),
    budget DECIMAL(15,2),
    status VARCHAR(30) NOT NULL DEFAULT 'new',
    next_follow_up_at TIMESTAMPTZ,
    notes TEXT,
    loss_reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leads_sales_id ON leads(sales_id);
CREATE INDEX idx_leads_status ON leads(status);
CREATE INDEX idx_leads_customer_id ON leads(customer_id);
CREATE INDEX idx_leads_property_id ON leads(property_id);
CREATE INDEX idx_leads_inquiry_id ON leads(inquiry_id);

CREATE TABLE IF NOT EXISTS lead_follow_ups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    sales_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(30) NOT NULL,
    note TEXT,
    follow_up_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lead_follow_ups_lead_id ON lead_follow_ups(lead_id);
CREATE INDEX idx_lead_follow_ups_sales_id ON lead_follow_ups(sales_id);

CREATE TABLE IF NOT EXISTS surveys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_id UUID NOT NULL REFERENCES leads(id) ON DELETE CASCADE,
    sales_id UUID NOT NULL REFERENCES users(id),
    scheduled_at TIMESTAMPTZ NOT NULL,
    location TEXT,
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'scheduled',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_surveys_lead_id ON surveys(lead_id);
CREATE INDEX idx_surveys_sales_id ON surveys(sales_id);
CREATE INDEX idx_surveys_status ON surveys(status);

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id),
    unit_id UUID REFERENCES units(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    emergency_contact VARCHAR(255),
    start_date DATE NOT NULL,
    end_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_property_id ON tenants(property_id);
CREATE INDEX idx_tenants_unit_id ON tenants(unit_id);
CREATE INDEX idx_tenants_status ON tenants(status);

CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    body TEXT,
    type VARCHAR(50) NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

DROP INDEX IF EXISTS idx_deals_outcome;
DROP INDEX IF EXISTS uq_deals_property_id;
ALTER TABLE deals DROP CONSTRAINT IF EXISTS chk_deals_deal_fields;
ALTER TABLE deals DROP CONSTRAINT IF EXISTS chk_deals_outcome;
DELETE FROM commissions WHERE deal_id IN (SELECT id FROM deals WHERE outcome = 'cancelled');
DELETE FROM deals WHERE outcome = 'cancelled';
ALTER TABLE deals ALTER COLUMN type SET NOT NULL;
ALTER TABLE deals ALTER COLUMN price SET NOT NULL;
-- lead_id cannot be restored (leads are gone); it comes back nullable.
ALTER TABLE deals ADD COLUMN lead_id UUID REFERENCES leads(id);
ALTER TABLE deals DROP COLUMN buyer_name;
ALTER TABLE deals DROP COLUMN outcome;
ALTER TABLE deals DROP COLUMN recorded_by;
