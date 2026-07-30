CREATE TABLE IF NOT EXISTS match_cycles (
    id BIGSERIAL PRIMARY KEY,
    cycle_key VARCHAR(64) NOT NULL,
    title VARCHAR(128) NOT NULL,
    reveal_status VARCHAR(16) NOT NULL DEFAULT 'draft',
    reveal_at TIMESTAMPTZ NOT NULL,
    participant_count INTEGER NOT NULL DEFAULT 0,
    matched_user_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (cycle_key)
);

CREATE INDEX IF NOT EXISTS idx_match_cycles_reveal_at
    ON match_cycles (reveal_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS user_match_results (
    id BIGSERIAL PRIMARY KEY,
    cycle_id BIGINT NOT NULL REFERENCES match_cycles(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id),
    result_status VARCHAR(16) NOT NULL,
    partner_user_id BIGINT REFERENCES users(id),
    match_score DOUBLE PRECISION,
    match_reason TEXT,
    result_message TEXT NOT NULL,
    partner_nickname VARCHAR(64),
    partner_department VARCHAR(64),
    partner_grade VARCHAR(64),
    partner_contacts_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    revealed_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (cycle_id, user_id),
    CHECK (result_status IN ('matched', 'failed')),
    CHECK (partner_user_id IS NULL OR partner_user_id <> user_id),
    CHECK (match_score IS NULL OR (match_score >= 0 AND match_score <= 100))
);

CREATE INDEX IF NOT EXISTS idx_user_match_results_user_id
    ON user_match_results (user_id, revealed_at DESC, id DESC);