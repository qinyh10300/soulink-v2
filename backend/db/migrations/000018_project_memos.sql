ALTER TABLE connection_intents
    ADD COLUMN IF NOT EXISTS project_memo JSONB NOT NULL DEFAULT '{}'::jsonb;

ALTER TABLE connection_intents
    DROP CONSTRAINT IF EXISTS connection_intents_project_memo_object;

ALTER TABLE connection_intents
    ADD CONSTRAINT connection_intents_project_memo_object
    CHECK (jsonb_typeof(project_memo) = 'object');
