ALTER TABLE safety_reports
    ADD COLUMN IF NOT EXISTS assigned_to BIGINT REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS review_notes TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

UPDATE safety_reports
SET updated_at = COALESCE(resolved_at, created_at);

ALTER TABLE safety_reports
    ADD CONSTRAINT safety_reports_review_notes_length
    CHECK (char_length(review_notes) <= 2000);

CREATE INDEX IF NOT EXISTS idx_safety_reports_assigned_status
    ON safety_reports (assigned_to, status, updated_at DESC, id DESC);
