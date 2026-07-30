CREATE TABLE IF NOT EXISTS user_questionnaire_portraits (
    user_id BIGINT NOT NULL REFERENCES users(id),
    questionnaire_key VARCHAR(128) NOT NULL,
    questionnaire_version VARCHAR(64) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    answer_signature TEXT NOT NULL DEFAULT '',
    soul_signature TEXT NOT NULL DEFAULT '',
    personal_profile TEXT NOT NULL DEFAULT '',
    ideal_resonance TEXT NOT NULL DEFAULT '',
    traits_json JSONB,
    error_message TEXT NOT NULL DEFAULT '',
    generated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, questionnaire_key, questionnaire_version)
);

CREATE INDEX IF NOT EXISTS idx_user_questionnaire_portraits_status
    ON user_questionnaire_portraits (questionnaire_key, questionnaire_version, status);
