CREATE TABLE IF NOT EXISTS media_assets (
    id          BIGSERIAL PRIMARY KEY,
    asset_key   UUID NOT NULL UNIQUE,
    owner_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind        VARCHAR(24) NOT NULL,
    filename    VARCHAR(255) NOT NULL DEFAULT '',
    mime_type   VARCHAR(64) NOT NULL,
    byte_size   INTEGER NOT NULL,
    width       INTEGER NOT NULL DEFAULT 0,
    height      INTEGER NOT NULL DEFAULT 0,
    is_animated BOOLEAN NOT NULL DEFAULT FALSE,
    content     BYTEA NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (kind IN ('plaza', 'avatar')),
    CHECK (byte_size > 0),
    CHECK (width >= 0 AND height >= 0)
);

CREATE INDEX IF NOT EXISTS idx_media_assets_owner_kind
    ON media_assets (owner_id, kind, created_at DESC);

CREATE TABLE IF NOT EXISTS playground_post_media (
    post_id    BIGINT NOT NULL REFERENCES playground_posts(id) ON DELETE CASCADE,
    media_id   BIGINT NOT NULL REFERENCES media_assets(id) ON DELETE CASCADE,
    sort_order SMALLINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (post_id, media_id),
    UNIQUE (media_id),
    CHECK (sort_order >= 0)
);

CREATE INDEX IF NOT EXISTS idx_playground_post_media_order
    ON playground_post_media (post_id, sort_order, media_id);

ALTER TABLE playground_posts
    ADD COLUMN IF NOT EXISTS favorite_count INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS playground_post_favorites (
    post_id    BIGINT NOT NULL REFERENCES playground_posts(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_playground_post_favorites_user
    ON playground_post_favorites (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS user_profile_settings (
    user_id             BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    profile_public      BOOLEAN NOT NULL DEFAULT FALSE,
    schedule_public     BOOLEAN NOT NULL DEFAULT FALSE,
    preferences_public  BOOLEAN NOT NULL DEFAULT FALSE,
    avatar_media_id     BIGINT REFERENCES media_assets(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_profile_settings_avatar
    ON user_profile_settings (avatar_media_id)
    WHERE avatar_media_id IS NOT NULL;
