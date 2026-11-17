ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_active ON refresh_tokens(user_id) WHERE revoked_at IS NULL;
