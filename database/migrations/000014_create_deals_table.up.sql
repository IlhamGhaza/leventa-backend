CREATE TABLE IF NOT EXISTS deals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_id UUID NOT NULL REFERENCES leads(id),
    property_id UUID NOT NULL REFERENCES properties(id),
    sales_id UUID NOT NULL REFERENCES users(id),
    type VARCHAR(20) NOT NULL,
    price DECIMAL(15,2) NOT NULL,
    deal_date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deals_lead_id ON deals(lead_id);
CREATE INDEX idx_deals_property_id ON deals(property_id);
CREATE INDEX idx_deals_sales_id ON deals(sales_id);
CREATE INDEX idx_deals_type ON deals(type);
