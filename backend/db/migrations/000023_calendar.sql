CREATE TABLE IF NOT EXISTS calendar_events (
    id           BIGSERIAL PRIMARY KEY,
    organizer_id BIGINT NOT NULL REFERENCES users(id),
    event_type   VARCHAR(20) NOT NULL DEFAULT 'meetup',
    title        VARCHAR(120) NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    location     VARCHAR(160) NOT NULL DEFAULT '',
    starts_at    TIMESTAMPTZ NOT NULL,
    ends_at      TIMESTAMPTZ NOT NULL,
    visibility   VARCHAR(20) NOT NULL DEFAULT 'busy_only',
    status       VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (event_type IN ('meetup', 'study', 'sports', 'meal', 'date', 'project', 'other')),
    CHECK (visibility IN ('public', 'busy_only', 'private')),
    CHECK (status IN ('active', 'cancelled')),
    CHECK (ends_at > starts_at)
);

CREATE INDEX IF NOT EXISTS idx_calendar_events_organizer_time
    ON calendar_events (organizer_id, starts_at, ends_at)
    WHERE status = 'active';

CREATE TABLE IF NOT EXISTS calendar_event_attendees (
    event_id     BIGINT NOT NULL REFERENCES calendar_events(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    responded_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, user_id),
    CHECK (status IN ('pending', 'accepted', 'declined'))
);

CREATE INDEX IF NOT EXISTS idx_calendar_attendees_user
    ON calendar_event_attendees (user_id, status, event_id);
