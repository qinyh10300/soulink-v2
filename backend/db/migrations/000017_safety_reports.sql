CREATE TABLE IF NOT EXISTS user_blocks (
    blocker_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason VARCHAR(240) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id),
    CONSTRAINT user_blocks_distinct_users CHECK (blocker_id <> blocked_id)
);

CREATE INDEX IF NOT EXISTS idx_user_blocks_blocked
    ON user_blocks (blocked_id, created_at DESC);

CREATE TABLE IF NOT EXISTS safety_reports (
    id BIGSERIAL PRIMARY KEY,
    reporter_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    intent_id BIGINT REFERENCES connection_intents(id) ON DELETE SET NULL,
    room_id BIGINT REFERENCES match_rooms(id) ON DELETE SET NULL,
    category VARCHAR(32) NOT NULL,
    details TEXT NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'open',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    resolved_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT safety_reports_category_check CHECK (
        category IN ('harassment', 'privacy', 'impersonation', 'unsafe', 'spam', 'other')
    ),
    CONSTRAINT safety_reports_status_check CHECK (
        status IN ('open', 'reviewing', 'resolved', 'dismissed')
    ),
    CONSTRAINT safety_reports_context_check CHECK (
        intent_id IS NOT NULL OR room_id IS NOT NULL
    )
);

CREATE INDEX IF NOT EXISTS idx_safety_reports_status
    ON safety_reports (status, created_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_safety_reports_reported_user
    ON safety_reports (reported_user_id, created_at DESC, id DESC);
