ALTER TABLE user_questionnaire_states
    ADD COLUMN IF NOT EXISTS participating_in_match BOOLEAN NOT NULL DEFAULT FALSE;
