CREATE TABLE IF NOT EXISTS matching_profiles (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    bio TEXT NOT NULL DEFAULT '',
    skills JSONB NOT NULL DEFAULT '[]'::jsonb,
    interests JSONB NOT NULL DEFAULT '[]'::jsonb,
    availability TEXT NOT NULL DEFAULT '',
    preferred_scenes JSONB NOT NULL DEFAULT '[]'::jsonb,
    collaboration_style TEXT NOT NULL DEFAULT '',
    boundaries TEXT NOT NULL DEFAULT '',
    ai_portrait TEXT NOT NULL DEFAULT '',
    ai_portrait_approved BOOLEAN NOT NULL DEFAULT FALSE,
    romantic_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    matching_paused BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT matching_profiles_skills_array CHECK (jsonb_typeof(skills) = 'array'),
    CONSTRAINT matching_profiles_interests_array CHECK (jsonb_typeof(interests) = 'array'),
    CONSTRAINT matching_profiles_scenes_array CHECK (jsonb_typeof(preferred_scenes) = 'array')
);

CREATE TABLE IF NOT EXISTS connection_intents (
    id BIGSERIAL PRIMARY KEY,
    author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(32) NOT NULL,
    title VARCHAR(80) NOT NULL,
    description TEXT NOT NULL,
    goal TEXT NOT NULL DEFAULT '',
    time_window VARCHAR(160) NOT NULL DEFAULT '',
    location VARCHAR(120) NOT NULL DEFAULT '',
    visibility VARCHAR(24) NOT NULL DEFAULT 'candidates',
    status VARCHAR(24) NOT NULL DEFAULT 'draft',
    constraints JSONB NOT NULL DEFAULT '[]'::jsonb,
    deal_breakers JSONB NOT NULL DEFAULT '[]'::jsonb,
    desired_next_step VARCHAR(200) NOT NULL DEFAULT '',
    expires_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT connection_intents_category_check CHECK (
        category IN ('meal', 'study', 'sport', 'event', 'conversation', 'romantic', 'project', 'cofounder', 'other')
    ),
    CONSTRAINT connection_intents_visibility_check CHECK (
        visibility IN ('private', 'candidates', 'feed')
    ),
    CONSTRAINT connection_intents_status_check CHECK (
        status IN ('draft', 'active', 'paused', 'fulfilled', 'expired')
    ),
    CONSTRAINT connection_intents_constraints_array CHECK (jsonb_typeof(constraints) = 'array'),
    CONSTRAINT connection_intents_deal_breakers_array CHECK (jsonb_typeof(deal_breakers) = 'array')
);

CREATE INDEX IF NOT EXISTS idx_connection_intents_feed
    ON connection_intents (category, published_at DESC, id DESC)
    WHERE status = 'active' AND visibility = 'feed';

CREATE INDEX IF NOT EXISTS idx_connection_intents_author
    ON connection_intents (author_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_connection_intents_recommendations
    ON connection_intents (category, updated_at DESC, id DESC)
    WHERE status = 'active' AND visibility IN ('candidates', 'feed');

CREATE TABLE IF NOT EXISTS intent_saves (
    intent_id BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (intent_id, user_id)
);

CREATE TABLE IF NOT EXISTS intent_interests (
    intent_id BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(24) NOT NULL DEFAULT 'interested',
    message VARCHAR(500) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (intent_id, user_id),
    CONSTRAINT intent_interests_status_check CHECK (
        status IN ('interested', 'accepted', 'declined', 'withdrawn')
    )
);

CREATE INDEX IF NOT EXISTS idx_intent_interests_status
    ON intent_interests (intent_id, status, updated_at DESC);

CREATE TABLE IF NOT EXISTS match_rooms (
    id BIGSERIAL PRIMARY KEY,
    intent_id BIGINT NOT NULL REFERENCES connection_intents(id) ON DELETE CASCADE,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    candidate_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(24) NOT NULL DEFAULT 'open',
    fit_reason TEXT NOT NULL DEFAULT '',
    uncertainty TEXT NOT NULL DEFAULT '',
    shared_context JSONB NOT NULL DEFAULT '[]'::jsonb,
    conversation_starters JSONB NOT NULL DEFAULT '[]'::jsonb,
    suggested_places JSONB NOT NULL DEFAULT '[]'::jsonb,
    scheduled_at TIMESTAMPTZ,
    selected_place VARCHAR(160) NOT NULL DEFAULT '',
    owner_contact_shared BOOLEAN NOT NULL DEFAULT FALSE,
    candidate_contact_shared BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (intent_id, candidate_id),
    CONSTRAINT match_rooms_distinct_users CHECK (owner_id <> candidate_id),
    CONSTRAINT match_rooms_status_check CHECK (status IN ('open', 'met', 'closed')),
    CONSTRAINT match_rooms_shared_context_array CHECK (jsonb_typeof(shared_context) = 'array'),
    CONSTRAINT match_rooms_starters_array CHECK (jsonb_typeof(conversation_starters) = 'array'),
    CONSTRAINT match_rooms_places_array CHECK (jsonb_typeof(suggested_places) = 'array')
);

CREATE INDEX IF NOT EXISTS idx_match_rooms_owner
    ON match_rooms (owner_id, updated_at DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_match_rooms_candidate
    ON match_rooms (candidate_id, updated_at DESC, id DESC);

CREATE TABLE IF NOT EXISTS match_room_messages (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES match_rooms(id) ON DELETE CASCADE,
    sender_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content VARCHAR(1000) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_match_room_messages_room
    ON match_room_messages (room_id, created_at ASC, id ASC);
