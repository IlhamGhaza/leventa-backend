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
