-- 好友关系
CREATE TABLE IF NOT EXISTS friendships (
    id           BIGSERIAL PRIMARY KEY,
    requester_id BIGINT NOT NULL REFERENCES users(id),
    addressee_id BIGINT NOT NULL REFERENCES users(id),
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(requester_id, addressee_id),
    CHECK(requester_id != addressee_id)
);
CREATE INDEX IF NOT EXISTS idx_friendships_addressee ON friendships(addressee_id, status);
CREATE INDEX IF NOT EXISTS idx_friendships_requester ON friendships(requester_id, status);

-- 一对一会话
CREATE TABLE IF NOT EXISTS conversations (
    id                 BIGSERIAL PRIMARY KEY,
    user1_id           BIGINT NOT NULL REFERENCES users(id),
    user2_id           BIGINT NOT NULL REFERENCES users(id),
    user1_last_read_at TIMESTAMPTZ,
    user2_last_read_at TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(user1_id, user2_id),
    CHECK(user1_id < user2_id)
);

-- 消息
CREATE TABLE IF NOT EXISTS messages (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id BIGINT NOT NULL REFERENCES conversations(id),
    sender_id       BIGINT NOT NULL REFERENCES users(id),
    content         TEXT NOT NULL,
    content_type    VARCHAR(20) NOT NULL DEFAULT 'text',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_messages_conv_time ON messages(conversation_id, created_at DESC);
