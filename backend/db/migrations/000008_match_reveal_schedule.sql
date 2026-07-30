ALTER TABLE user_questionnaire_states
    ADD COLUMN IF NOT EXISTS result_reveal_at TIMESTAMPTZ;

ALTER TABLE match_cycles
    ADD COLUMN IF NOT EXISTS questionnaire_key VARCHAR(64),
    ADD COLUMN IF NOT EXISTS questionnaire_version VARCHAR(64);

ALTER TABLE user_match_results
    ALTER COLUMN revealed_at DROP NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_questionnaire_states_reveal_at
    ON user_questionnaire_states (questionnaire_key, questionnaire_version, result_reveal_at);