CREATE TABLE IF NOT EXISTS product_events (
    id            BIGSERIAL PRIMARY KEY,
    event_name    VARCHAR(64) NOT NULL,
    actor_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    subject_type  VARCHAR(32) NOT NULL DEFAULT '',
    subject_id    BIGINT,
    request_id    VARCHAR(96) NOT NULL DEFAULT '',
    dedupe_key    VARCHAR(180),
    properties    JSONB NOT NULL DEFAULT '{}'::jsonb,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_events_name_check CHECK (
        event_name ~ '^[a-z][a-z0-9_]{2,63}$'
    ),
    CONSTRAINT product_events_subject_check CHECK (
        (subject_type = '' AND subject_id IS NULL)
        OR (subject_type <> '' AND subject_id IS NOT NULL AND subject_id > 0)
    ),
    CONSTRAINT product_events_properties_object CHECK (
        jsonb_typeof(properties) = 'object'
    )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_product_events_dedupe
    ON product_events (dedupe_key)
    WHERE dedupe_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_product_events_name_time
    ON product_events (event_name, occurred_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_product_events_actor_time
    ON product_events (actor_user_id, occurred_at DESC, id DESC)
    WHERE actor_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_product_events_subject
    ON product_events (subject_type, subject_id, occurred_at DESC)
    WHERE subject_type <> '';

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'intent_created', author_id, 'intent', id,
       'backfill:intent_created:' || id,
       jsonb_build_object('category', category, 'source', 'migration_backfill'),
       created_at
FROM connection_intents
ON CONFLICT DO NOTHING;

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'intent_published', author_id, 'intent', id,
       'backfill:intent_published:' || id,
       jsonb_build_object('category', category, 'source', 'migration_backfill'),
       COALESCE(published_at, updated_at)
FROM connection_intents
WHERE published_at IS NOT NULL OR status = 'active'
ON CONFLICT DO NOTHING;

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'interest_expressed', interest.user_id, 'intent', interest.intent_id,
       'backfill:interest_expressed:' || interest.intent_id || ':' || interest.user_id,
       jsonb_build_object('source', 'migration_backfill'),
       interest.created_at
FROM intent_interests interest
ON CONFLICT DO NOTHING;

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'interest_accepted', intent.author_id, 'intent', interest.intent_id,
       'backfill:interest_accepted:' || interest.intent_id || ':' || interest.user_id,
       jsonb_build_object('source', 'migration_backfill'),
       interest.updated_at
FROM intent_interests interest
JOIN connection_intents intent ON intent.id = interest.intent_id
WHERE interest.status = 'accepted'
ON CONFLICT DO NOTHING;

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'match_created', owner_id, 'match_room', id,
       'backfill:match_created:' || id,
       jsonb_build_object('source', 'migration_backfill'),
       created_at
FROM match_rooms
ON CONFLICT DO NOTHING;

INSERT INTO product_events (
    event_name, actor_user_id, subject_type, subject_id,
    dedupe_key, properties, occurred_at
)
SELECT 'safety_report_submitted', reporter_id, 'safety_report', id,
       'backfill:safety_report_submitted:' || id,
       jsonb_build_object('category', category, 'source', 'migration_backfill'),
       created_at
FROM safety_reports
ON CONFLICT DO NOTHING;
