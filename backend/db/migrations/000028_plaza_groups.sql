CREATE TABLE IF NOT EXISTS plaza_group_invites (
    id                    BIGSERIAL PRIMARY KEY,
    post_id               BIGINT NOT NULL UNIQUE REFERENCES playground_posts(id) ON DELETE CASCADE,
    owner_id              BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invite_type           VARCHAR(16) NOT NULL,
    title                 VARCHAR(100) NOT NULL,
    summary               VARCHAR(1000) NOT NULL DEFAULT '',
    extracted_time        VARCHAR(200) NOT NULL DEFAULT '',
    extracted_location    VARCHAR(160) NOT NULL DEFAULT '',
    min_members           SMALLINT NOT NULL DEFAULT 2,
    max_members           SMALLINT NOT NULL DEFAULT 6,
    reveal_strategy       VARCHAR(16) NOT NULL DEFAULT 'on_form',
    deadline              TIMESTAMPTZ NOT NULL,
    group_expires_at      TIMESTAMPTZ NOT NULL,
    status                VARCHAR(16) NOT NULL DEFAULT 'open',
    chat_group_id         BIGINT REFERENCES chat_groups(id) ON DELETE SET NULL,
    inference_source      VARCHAR(24) NOT NULL DEFAULT 'deterministic',
    inference_model       VARCHAR(120) NOT NULL DEFAULT '',
    inference_confidence  NUMERIC(4, 3) NOT NULL DEFAULT 0.58,
    prompt_version        VARCHAR(64) NOT NULL DEFAULT 'plaza-group-draft-v1',
    formed_at             TIMESTAMPTZ,
    archived_at           TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (invite_type IN ('event', 'group')),
    CHECK (min_members BETWEEN 2 AND 12),
    CHECK (max_members BETWEEN min_members AND 12),
    CHECK (reveal_strategy IN ('on_accept', 'on_form')),
    CHECK (status IN ('open', 'formed', 'expired', 'archived', 'kept', 'cancelled')),
    CHECK (inference_source IN ('model', 'deterministic')),
    CHECK (inference_confidence BETWEEN 0 AND 1),
    CHECK (deadline > created_at),
    CHECK (group_expires_at > deadline)
);

CREATE INDEX IF NOT EXISTS idx_plaza_group_invites_owner
    ON plaza_group_invites (owner_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_plaza_group_invites_lifecycle
    ON plaza_group_invites (status, deadline, group_expires_at);

CREATE TABLE IF NOT EXISTS plaza_group_applications (
    id                BIGSERIAL PRIMARY KEY,
    invite_id         BIGINT NOT NULL REFERENCES plaza_group_invites(id) ON DELETE CASCADE,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    comment_id        BIGINT NOT NULL REFERENCES playground_comments(id) ON DELETE CASCADE,
    anonymous_alias   VARCHAR(40) NOT NULL,
    status            VARCHAR(16) NOT NULL DEFAULT 'applied',
    responded_at      TIMESTAMPTZ,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (invite_id, user_id),
    CHECK (status IN ('applied', 'accepted', 'declined', 'waitlisted', 'left'))
);

CREATE INDEX IF NOT EXISTS idx_plaza_group_applications_invite
    ON plaza_group_applications (invite_id, status, created_at, id);

CREATE INDEX IF NOT EXISTS idx_plaza_group_applications_user
    ON plaza_group_applications (user_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS plaza_group_keep_votes (
    invite_id    BIGINT NOT NULL REFERENCES plaza_group_invites(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (invite_id, user_id)
);
