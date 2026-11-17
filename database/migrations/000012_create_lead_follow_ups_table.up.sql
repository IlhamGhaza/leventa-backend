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
