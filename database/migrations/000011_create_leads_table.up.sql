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
