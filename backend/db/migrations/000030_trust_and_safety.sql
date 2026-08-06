ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

UPDATE users
SET email_verified_at = created_at
WHERE email_verified_at IS NULL;

ALTER TABLE users
    ALTER COLUMN email_verified_at SET DEFAULT NOW(),
    ALTER COLUMN email_verified_at SET NOT NULL;

CREATE TABLE IF NOT EXISTS photo_verification_requests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    evidence_mime VARCHAR(32) NOT NULL,
    evidence_data BYTEA NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    review_notes TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    appeal_reason TEXT NOT NULL DEFAULT '',
    appealed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT photo_verification_mime_check CHECK (
        evidence_mime IN ('image/jpeg', 'image/png')
    ),
    CONSTRAINT photo_verification_status_check CHECK (
        status IN ('pending', 'approved', 'rejected')
    ),
    CONSTRAINT photo_verification_evidence_size CHECK (
        octet_length(evidence_data) BETWEEN 1 AND 2097152
    ),
    CONSTRAINT photo_verification_notes_length CHECK (
        char_length(review_notes) <= 1000 AND char_length(appeal_reason) <= 1000
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_photo_verification_pending_user
    ON photo_verification_requests (user_id)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_photo_verification_review_queue
    ON photo_verification_requests (status, updated_at DESC, id DESC);

ALTER TABLE safety_reports
    ADD COLUMN IF NOT EXISTS conversation_id BIGINT REFERENCES conversations(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS group_id BIGINT REFERENCES chat_groups(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS context_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS risk_signals JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE safety_reports
    DROP CONSTRAINT IF EXISTS safety_reports_context_check;

ALTER TABLE safety_reports
    ADD CONSTRAINT safety_reports_context_check CHECK (
        num_nonnulls(intent_id, room_id, conversation_id, group_id) = 1
    );

CREATE TABLE IF NOT EXISTS safety_report_appeals (
    id BIGSERIAL PRIMARY KEY,
    report_id BIGINT NOT NULL REFERENCES safety_reports(id) ON DELETE CASCADE,
    appellant_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    status VARCHAR(24) NOT NULL DEFAULT 'pending',
    review_notes TEXT NOT NULL DEFAULT '',
    reviewed_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT safety_report_appeal_reason_length CHECK (
        char_length(reason) BETWEEN 8 AND 1000
    ),
    CONSTRAINT safety_report_appeal_notes_length CHECK (
        char_length(review_notes) <= 1000
    ),
    CONSTRAINT safety_report_appeal_status_check CHECK (
        status IN ('pending', 'accepted', 'rejected')
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_safety_report_pending_appeal
    ON safety_report_appeals (report_id, appellant_id)
    WHERE status = 'pending';

CREATE INDEX IF NOT EXISTS idx_safety_report_appeals_queue
    ON safety_report_appeals (status, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS safety_review_events (
    id BIGSERIAL PRIMARY KEY,
    entity_type VARCHAR(32) NOT NULL,
    entity_id BIGINT NOT NULL,
    actor_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action VARCHAR(48) NOT NULL,
    previous_status VARCHAR(24) NOT NULL DEFAULT '',
    next_status VARCHAR(24) NOT NULL DEFAULT '',
    note TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT safety_review_event_entity_check CHECK (
        entity_type IN ('report', 'report_appeal', 'photo_verification')
    ),
    CONSTRAINT safety_review_event_note_length CHECK (char_length(note) <= 1000)
);

CREATE INDEX IF NOT EXISTS idx_safety_review_events_entity
    ON safety_review_events (entity_type, entity_id, created_at, id);

CREATE TABLE IF NOT EXISTS trusted_contacts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(64) NOT NULL,
    contact_type VARCHAR(24) NOT NULL,
    contact_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT trusted_contacts_type_check CHECK (
        contact_type IN ('email', 'phone', 'wechat', 'other')
    ),
    CONSTRAINT trusted_contacts_distinct_value UNIQUE (user_id, contact_type, contact_value)
);

CREATE INDEX IF NOT EXISTS idx_trusted_contacts_user
    ON trusted_contacts (user_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS meetup_safety_shares (
    id BIGSERIAL PRIMARY KEY,
    meetup_id BIGINT NOT NULL REFERENCES meetups(id) ON DELETE CASCADE,
    sharer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    trusted_contact_id BIGINT NOT NULL REFERENCES trusted_contacts(id) ON DELETE RESTRICT,
    summary JSONB NOT NULL,
    delivery_status VARCHAR(24) NOT NULL,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT meetup_safety_share_delivery_check CHECK (
        delivery_status IN ('sent', 'recorded', 'failed')
    )
);

CREATE INDEX IF NOT EXISTS idx_meetup_safety_shares_user
    ON meetup_safety_shares (sharer_id, created_at DESC, id DESC);
