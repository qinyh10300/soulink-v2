CREATE TABLE IF NOT EXISTS chat_groups (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(40) NOT NULL,
    owner_id   BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chat_groups_owner
    ON chat_groups (owner_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS chat_group_members (
    group_id     BIGINT NOT NULL REFERENCES chat_groups(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    role         VARCHAR(16) NOT NULL DEFAULT 'member',
    last_read_at TIMESTAMPTZ,
    joined_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    left_at      TIMESTAMPTZ,
    PRIMARY KEY (group_id, user_id),
    CHECK (role IN ('owner', 'member'))
);

CREATE INDEX IF NOT EXISTS idx_chat_group_members_user
    ON chat_group_members (user_id, joined_at DESC)
    WHERE left_at IS NULL;

CREATE TABLE IF NOT EXISTS chat_group_messages (
    id           BIGSERIAL PRIMARY KEY,
    group_id     BIGINT NOT NULL REFERENCES chat_groups(id) ON DELETE CASCADE,
    sender_id    BIGINT NOT NULL REFERENCES users(id),
    content      TEXT NOT NULL,
    content_type VARCHAR(20) NOT NULL DEFAULT 'text',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_group_messages_time
    ON chat_group_messages (group_id, created_at DESC, id DESC);
