DROP INDEX IF EXISTS idx_refresh_tokens_user_active;
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;
ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;
