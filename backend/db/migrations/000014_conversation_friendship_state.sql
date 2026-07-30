ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS friendship_disabled_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS friendship_disabled_by_user_id BIGINT REFERENCES users(id);
