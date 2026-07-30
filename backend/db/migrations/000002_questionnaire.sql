CREATE TABLE IF NOT EXISTS user_questionnaire_states (
    user_id BIGINT NOT NULL REFERENCES users(id),
    questionnaire_key VARCHAR(128) NOT NULL,
    questionnaire_title VARCHAR(255) NOT NULL,
    questionnaire_version VARCHAR(64) NOT NULL,
    total_count INTEGER NOT NULL,
    required_count INTEGER NOT NULL,
    answered_count INTEGER NOT NULL DEFAULT 0,
    required_answered_count INTEGER NOT NULL DEFAULT 0,
    progress_percent INTEGER NOT NULL DEFAULT 0,
    has_completed_questionnaire BOOLEAN NOT NULL DEFAULT FALSE,
    last_action VARCHAR(32) NOT NULL DEFAULT 'save',
    last_saved_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, questionnaire_key)
);

CREATE TABLE IF NOT EXISTS user_questionnaire_answers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    questionnaire_key VARCHAR(128) NOT NULL,
    questionnaire_version VARCHAR(64) NOT NULL,
    question_id VARCHAR(32) NOT NULL,
    question_type VARCHAR(32) NOT NULL,
    value_json JSONB NOT NULL,
    is_important BOOLEAN NOT NULL DEFAULT FALSE,
    answered_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    UNIQUE (user_id, questionnaire_key, question_id)
);

CREATE INDEX IF NOT EXISTS idx_user_questionnaire_answers_lookup
    ON user_questionnaire_answers (user_id, questionnaire_key);