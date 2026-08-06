CREATE TABLE IF NOT EXISTS group_match_intents (
    id                    BIGSERIAL PRIMARY KEY,
    organizer_id          BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category              VARCHAR(24) NOT NULL,
    title                 VARCHAR(100) NOT NULL,
    description           VARCHAR(1200) NOT NULL DEFAULT '',
    role_offer            VARCHAR(80) NOT NULL DEFAULT '',
    roles_needed          JSONB NOT NULL DEFAULT '[]'::jsonb,
    constraints           JSONB NOT NULL DEFAULT '[]'::jsonb,
    time_window           VARCHAR(200) NOT NULL DEFAULT '',
    location              VARCHAR(160) NOT NULL DEFAULT '',
    min_members           SMALLINT NOT NULL DEFAULT 3,
    max_members           SMALLINT NOT NULL DEFAULT 5,
    starts_at             TIMESTAMPTZ,
    ends_at               TIMESTAMPTZ,
    create_calendar_event BOOLEAN NOT NULL DEFAULT FALSE,
    status                VARCHAR(20) NOT NULL DEFAULT 'open',
    chat_group_id         BIGINT REFERENCES chat_groups(id) ON DELETE SET NULL,
    calendar_event_id     BIGINT REFERENCES calendar_events(id) ON DELETE SET NULL,
    group_intro           VARCHAR(1000) NOT NULL DEFAULT '',
    conversation_starters JSONB NOT NULL DEFAULT '[]'::jsonb,
    inference_source      VARCHAR(24) NOT NULL DEFAULT 'deterministic',
    inference_model       VARCHAR(120) NOT NULL DEFAULT '',
    inference_confidence  NUMERIC(4, 3) NOT NULL DEFAULT 0.72,
    prompt_version        VARCHAR(64) NOT NULL DEFAULT 'group-summary-v1',
    expires_at            TIMESTAMPTZ NOT NULL,
    formed_at             TIMESTAMPTZ,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (category IN ('meal', 'study', 'sport', 'event', 'project')),
    CHECK (jsonb_typeof(roles_needed) = 'array'),
    CHECK (jsonb_typeof(constraints) = 'array'),
    CHECK (jsonb_typeof(conversation_starters) = 'array'),
    CHECK (min_members BETWEEN 2 AND 12),
    CHECK (max_members BETWEEN min_members AND 12),
    CHECK ((starts_at IS NULL AND ends_at IS NULL) OR
           (starts_at IS NOT NULL AND ends_at IS NOT NULL AND ends_at > starts_at)),
    CHECK (create_calendar_event = FALSE OR starts_at IS NOT NULL),
    CHECK (status IN ('open', 'joined', 'formed', 'expired', 'cancelled')),
    CHECK (inference_source IN ('model', 'deterministic')),
    CHECK (inference_confidence BETWEEN 0 AND 1),
    CHECK (expires_at > created_at)
);

CREATE INDEX IF NOT EXISTS idx_group_match_intents_owner
    ON group_match_intents (organizer_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_group_match_intents_pool
    ON group_match_intents (category, expires_at, updated_at DESC, id DESC)
    WHERE status = 'open';

CREATE TABLE IF NOT EXISTS group_match_invitations (
    id                  BIGSERIAL PRIMARY KEY,
    group_intent_id     BIGINT NOT NULL REFERENCES group_match_intents(id) ON DELETE CASCADE,
    candidate_intent_id BIGINT NOT NULL REFERENCES group_match_intents(id) ON DELETE CASCADE,
    candidate_user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score               SMALLINT NOT NULL,
    fit_reason          VARCHAR(800) NOT NULL DEFAULT '',
    uncertainty         VARCHAR(500) NOT NULL DEFAULT '',
    shared_constraints  JSONB NOT NULL DEFAULT '[]'::jsonb,
    suggested_role      VARCHAR(80) NOT NULL DEFAULT '',
    status              VARCHAR(20) NOT NULL DEFAULT 'invited',
    responded_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_intent_id, candidate_user_id),
    CHECK (candidate_user_id <> 0),
    CHECK (score BETWEEN 0 AND 100),
    CHECK (jsonb_typeof(shared_constraints) = 'array'),
    CHECK (status IN ('invited', 'accepted', 'declined', 'waitlisted', 'left'))
);

CREATE INDEX IF NOT EXISTS idx_group_match_invitations_candidate
    ON group_match_invitations (candidate_user_id, status, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_group_match_invitations_group
    ON group_match_invitations (group_intent_id, status, created_at, id);
