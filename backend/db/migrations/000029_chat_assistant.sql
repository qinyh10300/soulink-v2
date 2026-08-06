CREATE TABLE IF NOT EXISTS chat_assistant_artifacts (
    id                 BIGSERIAL PRIMARY KEY,
    user_id            BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    context_type       VARCHAR(16) NOT NULL,
    context_id         BIGINT NOT NULL,
    artifact_type      VARCHAR(24) NOT NULL,
    payload            JSONB NOT NULL,
    message_cutoff_id  BIGINT NOT NULL DEFAULT 0,
    inference_source   VARCHAR(24) NOT NULL DEFAULT 'deterministic',
    inference_model    VARCHAR(120) NOT NULL DEFAULT '',
    confidence         NUMERIC(4, 3) NOT NULL DEFAULT 0.60,
    prompt_version     VARCHAR(64) NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at         TIMESTAMPTZ,
    CHECK (context_type IN ('direct', 'group')),
    CHECK (context_id > 0),
    CHECK (artifact_type IN ('starters', 'summary', 'actions', 'group_summary')),
    CHECK (jsonb_typeof(payload) = 'object'),
    CHECK (message_cutoff_id >= 0),
    CHECK (inference_source IN ('model', 'deterministic')),
    CHECK (confidence BETWEEN 0 AND 1)
);

CREATE INDEX IF NOT EXISTS idx_chat_assistant_artifacts_context
    ON chat_assistant_artifacts (
        user_id, context_type, context_id, artifact_type, created_at DESC, id DESC
    )
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_chat_assistant_artifacts_user
    ON chat_assistant_artifacts (user_id, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;
