ALTER TABLE users
ADD COLUMN IF NOT EXISTS school_key VARCHAR(16);

CREATE INDEX IF NOT EXISTS idx_users_school_key
ON users (school_key)
WHERE deleted_at IS NULL;
