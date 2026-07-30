ALTER TABLE user_questionnaire_portraits
    ADD COLUMN IF NOT EXISTS answer_signature TEXT NOT NULL DEFAULT '';
