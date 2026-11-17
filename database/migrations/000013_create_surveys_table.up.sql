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
