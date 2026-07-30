CREATE TABLE IF NOT EXISTS user_contact_methods (
    user_id BIGINT NOT NULL REFERENCES users(id),
    contact_type VARCHAR(32) NOT NULL,
    contact_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, contact_type)
);

CREATE INDEX IF NOT EXISTS idx_user_contact_methods_user_id
    ON user_contact_methods (user_id);

INSERT INTO user_contact_methods (user_id, contact_type, contact_value, created_at, updated_at)
SELECT id,
       contact_type,
       contact_value,
       COALESCE(created_at, NOW()),
       COALESCE(updated_at, created_at, NOW())
FROM users
WHERE deleted_at IS NULL
  AND contact_type IS NOT NULL
  AND contact_type <> ''
  AND contact_value IS NOT NULL
  AND contact_value <> ''
ON CONFLICT (user_id, contact_type)
DO UPDATE SET
    contact_value = EXCLUDED.contact_value,
    updated_at = EXCLUDED.updated_at;