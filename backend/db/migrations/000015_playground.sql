CREATE TABLE IF NOT EXISTS playground_posts (
    id            BIGSERIAL PRIMARY KEY,
    author_id     BIGINT NOT NULL REFERENCES users(id),
    content       TEXT NOT NULL,
    image_urls    JSONB NOT NULL DEFAULT '[]'::jsonb,
    like_count    INTEGER NOT NULL DEFAULT 0,
    comment_count INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at    TIMESTAMPTZ,
    CHECK (jsonb_typeof(image_urls) = 'array')
);

CREATE INDEX IF NOT EXISTS idx_playground_posts_feed
    ON playground_posts (created_at DESC, id DESC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_playground_posts_author
    ON playground_posts (author_id, created_at DESC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS playground_post_likes (
    post_id    BIGINT NOT NULL REFERENCES playground_posts(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_playground_post_likes_user
    ON playground_post_likes (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS playground_comments (
    id                BIGSERIAL PRIMARY KEY,
    post_id           BIGINT NOT NULL REFERENCES playground_posts(id) ON DELETE CASCADE,
    author_id         BIGINT NOT NULL REFERENCES users(id),
    parent_comment_id BIGINT REFERENCES playground_comments(id) ON DELETE CASCADE,
    content           TEXT NOT NULL,
    like_count        INTEGER NOT NULL DEFAULT 0,
    reply_count       INTEGER NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_playground_comments_post
    ON playground_comments (post_id, created_at ASC, id ASC)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_playground_comments_parent
    ON playground_comments (parent_comment_id, created_at ASC, id ASC)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS playground_comment_likes (
    comment_id BIGINT NOT NULL REFERENCES playground_comments(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_playground_comment_likes_user
    ON playground_comment_likes (user_id, created_at DESC);
