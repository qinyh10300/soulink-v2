ALTER TABLE matching_profiles
    ADD COLUMN IF NOT EXISTS hidden_categories JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE matching_profiles
    DROP CONSTRAINT IF EXISTS matching_profiles_hidden_categories_array;

ALTER TABLE matching_profiles
    ADD CONSTRAINT matching_profiles_hidden_categories_array
    CHECK (jsonb_typeof(hidden_categories) = 'array');

CREATE TABLE IF NOT EXISTS intent_dismissals (
    intent_id BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (intent_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_intent_dismissals_user
    ON intent_dismissals (user_id, created_at DESC);
