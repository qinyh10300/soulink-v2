CREATE TABLE IF NOT EXISTS user_feedback (
  id BIGSERIAL PRIMARY KEY,
  ticket_id VARCHAR(64) NOT NULL UNIQUE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  user_email VARCHAR(255) NOT NULL,
  category VARCHAR(32) NOT NULL,
  rating SMALLINT NOT NULL,
  details TEXT NOT NULL,
  contact_email VARCHAR(255),
  submitted_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_feedback_user_id_submitted_at
ON user_feedback (user_id, submitted_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_feedback_submitted_at
ON user_feedback (submitted_at DESC);
