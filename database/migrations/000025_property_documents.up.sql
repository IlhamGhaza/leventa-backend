-- Ownership documents per property type (D-029). A document file may be stored
-- only when the matching cert_* flag on the property is TRUE.
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_shmsrs BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_hgu BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS cert_slf BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS property_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    doc_type VARCHAR(20) NOT NULL,
    storage_path VARCHAR(500) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    uploaded_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_property_documents_type UNIQUE (property_id, doc_type),
    CONSTRAINT chk_property_documents_type CHECK (doc_type IN ('shm', 'hgb', 'hgu', 'shmsrs', 'imb', 'pbb', 'girik', 'slf', 'other'))
);

CREATE INDEX IF NOT EXISTS idx_property_documents_property_id ON property_documents(property_id);
