ALTER TABLE user_questionnaire_states
    DROP CONSTRAINT IF EXISTS user_questionnaire_states_pkey;

ALTER TABLE user_questionnaire_states
    ADD CONSTRAINT user_questionnaire_states_pkey
    PRIMARY KEY (user_id, questionnaire_key, questionnaire_version);

ALTER TABLE user_questionnaire_answers
    DROP CONSTRAINT IF EXISTS user_questionnaire_answers_user_id_questionnaire_key_question_id_key;

ALTER TABLE user_questionnaire_answers
    DROP CONSTRAINT IF EXISTS user_questionnaire_answers_user_key_version_question_key;

ALTER TABLE user_questionnaire_answers
    ADD CONSTRAINT user_questionnaire_answers_user_key_version_question_key
    UNIQUE (user_id, questionnaire_key, questionnaire_version, question_id);

DROP INDEX IF EXISTS idx_user_questionnaire_answers_lookup;

CREATE INDEX IF NOT EXISTS idx_user_questionnaire_answers_lookup_version
    ON user_questionnaire_answers (user_id, questionnaire_key, questionnaire_version, answered_at, id);