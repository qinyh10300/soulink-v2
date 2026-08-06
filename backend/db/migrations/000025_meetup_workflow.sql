CREATE TABLE IF NOT EXISTS availability_windows (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    window_type  VARCHAR(16) NOT NULL,
    timezone     VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
    starts_at    TIMESTAMPTZ,
    ends_at      TIMESTAMPTZ,
    weekday      SMALLINT,
    start_minute SMALLINT,
    end_minute   SMALLINT,
    valid_from   DATE,
    valid_until  DATE,
    enabled      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (window_type IN ('one_time', 'weekly')),
    CHECK (
        (window_type = 'one_time'
            AND starts_at IS NOT NULL AND ends_at IS NOT NULL AND ends_at > starts_at
            AND weekday IS NULL AND start_minute IS NULL AND end_minute IS NULL)
        OR
        (window_type = 'weekly'
            AND starts_at IS NULL AND ends_at IS NULL
            AND weekday BETWEEN 1 AND 7
            AND start_minute BETWEEN 0 AND 1439
            AND end_minute BETWEEN 1 AND 1440
            AND end_minute > start_minute
            AND (valid_until IS NULL OR valid_from IS NULL OR valid_until >= valid_from))
    )
);

CREATE INDEX IF NOT EXISTS idx_availability_windows_user
    ON availability_windows (user_id, enabled, window_type, id);

CREATE INDEX IF NOT EXISTS idx_availability_windows_one_time
    ON availability_windows (user_id, starts_at, ends_at)
    WHERE enabled = TRUE AND window_type = 'one_time';

CREATE TABLE IF NOT EXISTS meetups (
    id                BIGSERIAL PRIMARY KEY,
    context_type      VARCHAR(20) NOT NULL,
    context_id        BIGINT NOT NULL,
    organizer_id      BIGINT NOT NULL REFERENCES users(id),
    title             VARCHAR(120) NOT NULL,
    category          VARCHAR(24) NOT NULL DEFAULT 'other',
    starts_at         TIMESTAMPTZ NOT NULL,
    ends_at           TIMESTAMPTZ NOT NULL,
    location          VARCHAR(160) NOT NULL DEFAULT '',
    status            VARCHAR(24) NOT NULL DEFAULT 'proposed',
    version           INTEGER NOT NULL DEFAULT 1,
    calendar_event_id BIGINT REFERENCES calendar_events(id) ON DELETE SET NULL,
    cancelled_by      BIGINT REFERENCES users(id),
    cancelled_reason  VARCHAR(240) NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (context_type IN ('match_room', 'direct', 'group')),
    CHECK (category IN ('meal', 'study', 'sport', 'event', 'romantic', 'project', 'other')),
    CHECK (status IN (
        'proposed', 'confirmed', 'declined', 'reschedule_requested',
        'cancelled', 'completed', 'no_show'
    )),
    CHECK (ends_at > starts_at),
    CHECK (version > 0)
);

CREATE INDEX IF NOT EXISTS idx_meetups_context
    ON meetups (context_type, context_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_meetups_time
    ON meetups (starts_at, status, id);

CREATE TABLE IF NOT EXISTS meetup_participants (
    meetup_id   BIGINT NOT NULL REFERENCES meetups(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    responded_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (meetup_id, user_id),
    CHECK (status IN ('pending', 'accepted', 'declined'))
);

CREATE INDEX IF NOT EXISTS idx_meetup_participants_user
    ON meetup_participants (user_id, status, meetup_id);

CREATE TABLE IF NOT EXISTS meetup_feedback (
    meetup_id             BIGINT NOT NULL REFERENCES meetups(id) ON DELETE CASCADE,
    user_id               BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attendance            VARCHAR(16) NOT NULL,
    rating                SMALLINT,
    would_meet_again      BOOLEAN,
    safety_concern        BOOLEAN NOT NULL DEFAULT FALSE,
    private_notes         VARCHAR(1000) NOT NULL DEFAULT '',
    submitted_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at            TIMESTAMPTZ,
    PRIMARY KEY (meetup_id, user_id),
    CHECK (attendance IN ('attended', 'no_show', 'cancelled')),
    CHECK (rating IS NULL OR rating BETWEEN 1 AND 5)
);

CREATE INDEX IF NOT EXISTS idx_meetup_feedback_active
    ON meetup_feedback (user_id, submitted_at DESC)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS recommendation_learning_signals (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    other_user_id     BIGINT REFERENCES users(id) ON DELETE CASCADE,
    source_meetup_id  BIGINT NOT NULL REFERENCES meetups(id) ON DELETE CASCADE,
    signal_type       VARCHAR(32) NOT NULL,
    weight            SMALLINT NOT NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    revoked_at        TIMESTAMPTZ,
    UNIQUE (user_id, other_user_id, source_meetup_id, signal_type),
    CHECK (signal_type IN ('meet_again', 'do_not_repeat', 'positive_meetup', 'no_show')),
    CHECK (weight BETWEEN -5 AND 5)
);

CREATE INDEX IF NOT EXISTS idx_recommendation_learning_signals_user
    ON recommendation_learning_signals (user_id, other_user_id, created_at DESC)
    WHERE revoked_at IS NULL;
