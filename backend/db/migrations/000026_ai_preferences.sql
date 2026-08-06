CREATE TABLE IF NOT EXISTS ai_user_preferences (
    user_id                         BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    recommendation_learning_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    assistant_enabled               BOOLEAN NOT NULL DEFAULT TRUE,
    learning_reset_at               TIMESTAMPTZ,
    assistant_memory_reset_at       TIMESTAMPTZ,
    created_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS recommendation_interaction_signals (
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    intent_id     BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    other_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category      VARCHAR(24) NOT NULL,
    signal_type   VARCHAR(32) NOT NULL,
    weight        SMALLINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    revoked_at    TIMESTAMPTZ,
    UNIQUE (user_id, intent_id, signal_type),
    CHECK (category IN (
        'meal', 'study', 'sport', 'event', 'conversation',
        'romantic', 'project', 'cofounder', 'other'
    )),
    CHECK (signal_type IN (
        'saved', 'interested', 'withdrawn', 'dismissed',
        'accepted_candidate', 'declined_candidate'
    )),
    CHECK (weight BETWEEN -5 AND 5),
    CHECK (other_user_id <> user_id)
);

CREATE INDEX IF NOT EXISTS idx_recommendation_interaction_signals_user
    ON recommendation_interaction_signals (user_id, created_at DESC, id DESC)
    WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_recommendation_interaction_signals_affinity
    ON recommendation_interaction_signals (user_id, other_user_id, category)
    WHERE revoked_at IS NULL;

CREATE TABLE IF NOT EXISTS daily_recommendations (
    id                       BIGSERIAL PRIMARY KEY,
    user_id                  BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    source_intent_id         BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    candidate_intent_id      BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    rank                     SMALLINT NOT NULL,
    score                    SMALLINT NOT NULL,
    fit_reason               VARCHAR(800) NOT NULL,
    uncertainty              VARCHAR(500) NOT NULL,
    shared_context           JSONB NOT NULL DEFAULT '[]'::jsonb,
    algorithm_version        VARCHAR(64) NOT NULL,
    inference_source         VARCHAR(24) NOT NULL,
    model_name               VARCHAR(120) NOT NULL DEFAULT '',
    confidence               NUMERIC(4, 3) NOT NULL,
    prompt_version           VARCHAR(64) NOT NULL,
    source_intent_updated_at TIMESTAMPTZ NOT NULL,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at               TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, source_intent_id, created_at, rank),
    CHECK (rank BETWEEN 1 AND 3),
    CHECK (score BETWEEN 0 AND 100),
    CHECK (jsonb_typeof(shared_context) = 'array'),
    CHECK (inference_source IN ('model', 'deterministic')),
    CHECK (confidence BETWEEN 0 AND 1),
    CHECK (expires_at > created_at),
    CHECK (candidate_intent_id <> source_intent_id)
);

CREATE INDEX IF NOT EXISTS idx_daily_recommendations_active
    ON daily_recommendations (user_id, source_intent_id, expires_at DESC, rank);

CREATE INDEX IF NOT EXISTS idx_daily_recommendations_freshness
    ON daily_recommendations (user_id, candidate_intent_id, created_at DESC);

ALTER TABLE match_rooms
    ADD COLUMN IF NOT EXISTS assistant_source VARCHAR(24) NOT NULL DEFAULT 'deterministic',
    ADD COLUMN IF NOT EXISTS assistant_model VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS assistant_confidence NUMERIC(4, 3) NOT NULL DEFAULT 0.72,
    ADD COLUMN IF NOT EXISTS assistant_prompt_version VARCHAR(64) NOT NULL DEFAULT 'conversation-starters-v1';

ALTER TABLE match_rooms
    DROP CONSTRAINT IF EXISTS match_rooms_assistant_source_check;

ALTER TABLE match_rooms
    ADD CONSTRAINT match_rooms_assistant_source_check
    CHECK (assistant_source IN ('model', 'deterministic'));

ALTER TABLE match_rooms
    DROP CONSTRAINT IF EXISTS match_rooms_assistant_confidence_check;

ALTER TABLE match_rooms
    ADD CONSTRAINT match_rooms_assistant_confidence_check
    CHECK (assistant_confidence BETWEEN 0 AND 1);

INSERT INTO recommendation_interaction_signals (
    user_id, intent_id, other_user_id, category, signal_type,
    weight, created_at, updated_at
)
SELECT interest.user_id, interest.intent_id, intent.author_id, intent.category,
       CASE WHEN interest.status = 'withdrawn' THEN 'withdrawn' ELSE 'interested' END,
       CASE WHEN interest.status = 'withdrawn' THEN -1 ELSE 3 END,
       interest.created_at, interest.updated_at
FROM intent_interests interest
JOIN connection_intents intent ON intent.id = interest.intent_id
WHERE interest.user_id <> intent.author_id
ON CONFLICT (user_id, intent_id, signal_type) DO NOTHING;

INSERT INTO recommendation_interaction_signals (
    user_id, intent_id, other_user_id, category, signal_type,
    weight, created_at, updated_at
)
SELECT dismissal.user_id, dismissal.intent_id, intent.author_id, intent.category,
       'dismissed', -2, dismissal.created_at, dismissal.created_at
FROM intent_dismissals dismissal
JOIN connection_intents intent ON intent.id = dismissal.intent_id
WHERE dismissal.user_id <> intent.author_id
ON CONFLICT (user_id, intent_id, signal_type) DO NOTHING;
