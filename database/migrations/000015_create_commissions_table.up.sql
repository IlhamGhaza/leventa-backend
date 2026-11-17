CREATE TABLE IF NOT EXISTS commissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    deal_id UUID NOT NULL REFERENCES deals(id),
    sales_id UUID NOT NULL REFERENCES users(id),
    percentage DECIMAL(5,2) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_commissions_deal_id ON commissions(deal_id);
CREATE INDEX idx_commissions_sales_id ON commissions(sales_id);
CREATE INDEX idx_commissions_status ON commissions(status);
