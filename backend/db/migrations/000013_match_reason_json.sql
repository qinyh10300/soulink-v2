ALTER TABLE user_match_results
ADD COLUMN IF NOT EXISTS match_reason_json JSONB;

UPDATE user_match_results
SET match_reason_json = jsonb_build_object('text', match_reason)
WHERE match_reason_json IS NULL
  AND NULLIF(BTRIM(match_reason), '') IS NOT NULL;
