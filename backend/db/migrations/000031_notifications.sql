CREATE TABLE IF NOT EXISTS user_engagement_preferences (
    user_id                    BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    notifications_enabled     BOOLEAN NOT NULL DEFAULT FALSE,
    new_messages              BOOLEAN NOT NULL DEFAULT TRUE,
    meetup_invites            BOOLEAN NOT NULL DEFAULT TRUE,
    meetup_reminders          BOOLEAN NOT NULL DEFAULT TRUE,
    daily_picks               BOOLEAN NOT NULL DEFAULT TRUE,
    permission_state          VARCHAR(16) NOT NULL DEFAULT 'default',
    timezone                  VARCHAR(64) NOT NULL DEFAULT 'Asia/Shanghai',
    selected_scenes           JSONB NOT NULL DEFAULT '[]'::jsonb,
    onboarding_completed_at   TIMESTAMPTZ,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_engagement_permission_check CHECK (
        permission_state IN ('default', 'granted', 'denied', 'unavailable')
    ),
    CONSTRAINT user_engagement_scenes_array CHECK (
        jsonb_typeof(selected_scenes) = 'array'
    )
);

CREATE TABLE IF NOT EXISTS campus_themes (
    theme_key       VARCHAR(48) PRIMARY KEY,
    title           VARCHAR(48) NOT NULL,
    description     VARCHAR(240) NOT NULL,
    category        VARCHAR(24) NOT NULL,
    default_prompt  VARCHAR(500) NOT NULL,
    schedule_hint   VARCHAR(80) NOT NULL DEFAULT '',
    icon_name       VARCHAR(32) NOT NULL DEFAULT 'sparkles',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    display_order   SMALLINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT campus_themes_key_check CHECK (
        theme_key ~ '^[a-z][a-z0-9-]{2,47}$'
    ),
    CONSTRAINT campus_themes_category_check CHECK (
        category IN ('meal', 'study', 'sport', 'event', 'romantic', 'project')
    ),
    CONSTRAINT campus_themes_order_check CHECK (
        display_order BETWEEN 0 AND 1000
    )
);

CREATE INDEX IF NOT EXISTS idx_campus_themes_active
    ON campus_themes (is_active DESC, display_order, theme_key);

CREATE TABLE IF NOT EXISTS user_campus_themes (
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme_key    VARCHAR(48) NOT NULL REFERENCES campus_themes(theme_key) ON DELETE CASCADE,
    selected_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, theme_key)
);

CREATE INDEX IF NOT EXISTS idx_user_campus_themes_theme
    ON user_campus_themes (theme_key, selected_at DESC, user_id);

CREATE TABLE IF NOT EXISTS user_notifications (
    id             BIGSERIAL PRIMARY KEY,
    user_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind           VARCHAR(32) NOT NULL,
    title          VARCHAR(100) NOT NULL,
    body           VARCHAR(240) NOT NULL,
    action_path    VARCHAR(240) NOT NULL,
    source_type    VARCHAR(32) NOT NULL DEFAULT '',
    source_id      BIGINT,
    dedupe_key     VARCHAR(180) NOT NULL,
    scheduled_for  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    delivered_at   TIMESTAMPTZ,
    read_at        TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT user_notifications_kind_check CHECK (
        kind IN ('new_message', 'meetup_invite', 'meetup_reminder', 'daily_picks')
    ),
    CONSTRAINT user_notifications_source_check CHECK (
        (source_type = '' AND source_id IS NULL)
        OR (source_type <> '' AND source_id IS NOT NULL AND source_id > 0)
    ),
    CONSTRAINT user_notifications_action_path_check CHECK (
        action_path ~ '^/[a-zA-Z0-9/_?=&.-]*$'
    ),
    CONSTRAINT user_notifications_dedupe UNIQUE (user_id, dedupe_key)
);

CREATE INDEX IF NOT EXISTS idx_user_notifications_inbox
    ON user_notifications (user_id, scheduled_for DESC, id DESC);

CREATE INDEX IF NOT EXISTS idx_user_notifications_unread
    ON user_notifications (user_id, id DESC)
    WHERE read_at IS NULL;

INSERT INTO campus_themes (
    theme_key, title, description, category, default_prompt,
    schedule_hint, icon_name, display_order
) VALUES
    (
        'tonight-meal', '今晚饭搭子', '今晚在校内找一位时间合拍、话题相投的同学。',
        'meal', '今晚想在紫荆附近找一位饭搭子，时间和餐厅都可以一起商量。',
        '今天 17:00 后', 'utensils', 10
    ),
    (
        'week-study', '本周自习', '按共同空闲和学习节奏，约一次专注自习。',
        'study', '这周想找一位安静高效的自习搭子，一起安排两小时专注学习。',
        '未来 7 天', 'book-open', 20
    ),
    (
        'weekend-sport', '周末运动', '跑步、球类或健身，从一场轻量活动开始。',
        'sport', '这个周末想在校内找运动搭子，项目和强度都可以先聊。',
        '本周末', 'dumbbell', 30
    ),
    (
        'find-teammates', '寻找队友', '围绕真实项目补齐角色，先完成一次短聊。',
        'project', '正在做一个校园项目，希望找到能力互补的队友，先约 30 分钟对齐目标。',
        '长期开放', 'users-round', 40
    )
ON CONFLICT (theme_key) DO NOTHING;

CREATE OR REPLACE FUNCTION soulink_notify_direct_message()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    recipient_id BIGINT;
BEGIN
    SELECT CASE
        WHEN conversation.user1_id = NEW.sender_id THEN conversation.user2_id
        ELSE conversation.user1_id
    END
    INTO recipient_id
    FROM conversations conversation
    WHERE conversation.id = NEW.conversation_id;

    INSERT INTO user_notifications (
        user_id, kind, title, body, action_path, source_type, source_id,
        dedupe_key, scheduled_for, created_at
    )
    SELECT recipient_id, 'new_message', '你收到一条新消息',
           '打开 SouLink 查看，系统通知不会显示聊天正文。',
           '/chat/direct/' || NEW.conversation_id,
           'conversation', NEW.conversation_id,
           'direct-message:' || NEW.id, NEW.created_at, NEW.created_at
    FROM user_engagement_preferences preference
    WHERE preference.user_id = recipient_id
      AND preference.notifications_enabled
      AND preference.new_messages
    ON CONFLICT (user_id, dedupe_key) DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_soulink_notify_direct_message ON messages;
CREATE TRIGGER trg_soulink_notify_direct_message
    AFTER INSERT ON messages
    FOR EACH ROW EXECUTE FUNCTION soulink_notify_direct_message();

CREATE OR REPLACE FUNCTION soulink_notify_group_message()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO user_notifications (
        user_id, kind, title, body, action_path, source_type, source_id,
        dedupe_key, scheduled_for, created_at
    )
    SELECT member.user_id, 'new_message', '群聊里有一条新消息',
           '打开 SouLink 查看，系统通知不会显示聊天正文。',
           '/chat/group/' || NEW.group_id,
           'group', NEW.group_id,
           'group-message:' || NEW.id, NEW.created_at, NEW.created_at
    FROM chat_group_members member
    JOIN user_engagement_preferences preference ON preference.user_id = member.user_id
    WHERE member.group_id = NEW.group_id
      AND member.left_at IS NULL
      AND member.user_id <> NEW.sender_id
      AND preference.notifications_enabled
      AND preference.new_messages
    ON CONFLICT (user_id, dedupe_key) DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_soulink_notify_group_message ON chat_group_messages;
CREATE TRIGGER trg_soulink_notify_group_message
    AFTER INSERT ON chat_group_messages
    FOR EACH ROW EXECUTE FUNCTION soulink_notify_group_message();

CREATE OR REPLACE FUNCTION soulink_notify_meetup_invite()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO user_notifications (
        user_id, kind, title, body, action_path, source_type, source_id,
        dedupe_key, scheduled_for, created_at
    )
    SELECT NEW.user_id, 'meetup_invite', '你收到一项约见邀请',
           '时间和地点仅在登录 SouLink 后显示。',
           '/calendar', 'meetup', meetup.id,
           'meetup-invite:' || meetup.id || ':1:' || NEW.user_id,
           NEW.created_at, NEW.created_at
    FROM meetups meetup
    JOIN user_engagement_preferences preference ON preference.user_id = NEW.user_id
    WHERE meetup.id = NEW.meetup_id
      AND NEW.status = 'pending'
      AND NEW.user_id <> meetup.organizer_id
      AND preference.notifications_enabled
      AND preference.meetup_invites
    ON CONFLICT (user_id, dedupe_key) DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_soulink_notify_meetup_invite ON meetup_participants;
CREATE TRIGGER trg_soulink_notify_meetup_invite
    AFTER INSERT ON meetup_participants
    FOR EACH ROW EXECUTE FUNCTION soulink_notify_meetup_invite();

CREATE OR REPLACE FUNCTION soulink_notify_meetup_reschedule()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF NEW.status <> 'reschedule_requested'
       OR NEW.status IS NOT DISTINCT FROM OLD.status
       OR NEW.version IS NOT DISTINCT FROM OLD.version THEN
        RETURN NEW;
    END IF;

    INSERT INTO user_notifications (
        user_id, kind, title, body, action_path, source_type, source_id,
        dedupe_key, scheduled_for, created_at
    )
    SELECT participant.user_id, 'meetup_invite', '一项约见等待你确认新时间',
           '时间和地点仅在登录 SouLink 后显示。',
           '/calendar', 'meetup', NEW.id,
           'meetup-reschedule:' || NEW.id || ':' || NEW.version || ':' || participant.user_id,
           NEW.updated_at, NEW.updated_at
    FROM meetup_participants participant
    JOIN user_engagement_preferences preference ON preference.user_id = participant.user_id
    WHERE participant.meetup_id = NEW.id
      AND participant.status = 'pending'
      AND participant.user_id <> NEW.organizer_id
      AND preference.notifications_enabled
      AND preference.meetup_invites
    ON CONFLICT (user_id, dedupe_key) DO NOTHING;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_soulink_notify_meetup_reschedule ON meetups;
CREATE TRIGGER trg_soulink_notify_meetup_reschedule
    AFTER UPDATE ON meetups
    FOR EACH ROW EXECUTE FUNCTION soulink_notify_meetup_reschedule();
