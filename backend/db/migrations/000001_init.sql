CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    nickname VARCHAR(64) NOT NULL,
    department VARCHAR(64) NOT NULL,
    grade VARCHAR(64) NOT NULL,
    contact_type VARCHAR(32),
    contact_value VARCHAR(128),
    has_completed_questionnaire BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email_active
    ON users (email)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS verification_codes (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL,
    purpose VARCHAR(32) NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    sent_at TIMESTAMPTZ NOT NULL,
    verified_at TIMESTAMPTZ,
    consumed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_verification_codes_lookup
    ON verification_codes (email, purpose, sent_at DESC);

CREATE TABLE IF NOT EXISTS auth_sessions (
    id UUID PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_auth_sessions_user_id
    ON auth_sessions (user_id);
