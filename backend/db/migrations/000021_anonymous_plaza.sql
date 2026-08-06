ALTER TABLE playground_posts
    ADD COLUMN IF NOT EXISTS topic VARCHAR(24) NOT NULL DEFAULT 'campus';

ALTER TABLE playground_posts
    DROP CONSTRAINT IF EXISTS playground_posts_topic_check;

ALTER TABLE playground_posts
    ADD CONSTRAINT playground_posts_topic_check
    CHECK (topic IN ('campus', 'partner', 'study', 'emotion', 'secondhand'));

CREATE INDEX IF NOT EXISTS idx_playground_posts_topic_feed
    ON playground_posts (topic, created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_playground_posts_content_search
    ON playground_posts USING gin (to_tsvector('simple', content))
    WHERE deleted_at IS NULL;
