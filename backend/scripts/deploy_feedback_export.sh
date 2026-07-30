#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVER_HOST="${SERVER_HOST:-203.0.113.10}"
SERVER_USER="${SERVER_USER:-root}"
SERVER_PORT="${SERVER_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/opt/soulink-feedback-export}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-soulink-postgres}"
POSTGRES_USER="${POSTGRES_USER:-solink}"
POSTGRES_DB="${POSTGRES_DB:-soulink-web}"
FEEDBACK_SCHEMA="${FEEDBACK_SCHEMA:-public}"
SINCE="${SINCE:-7d}"
OUTPUT_FILE="${OUTPUT_FILE:-feedback.csv}"

cd "${ROOT_DIR}"

echo "[1/4] Preparing remote directory on ${SERVER_USER}@${SERVER_HOST}..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p '${REMOTE_DIR}'"

echo "[2/4] Exporting feedback rows from remote PostgreSQL container..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "\
docker exec '${POSTGRES_CONTAINER}' psql -U '${POSTGRES_USER}' -d '${POSTGRES_DB}' -c \"\\copy (
SELECT
  ticket_id,
  user_id,
  user_email,
  category,
  rating,
  details,
  contact_email,
  submitted_at
FROM ${FEEDBACK_SCHEMA}.user_feedback
WHERE submitted_at >= NOW() - INTERVAL '${SINCE}'
ORDER BY submitted_at DESC
) TO STDOUT WITH CSV HEADER\" > '${REMOTE_DIR}/${OUTPUT_FILE}'"

echo "[3/4] Downloading ${OUTPUT_FILE}..."
scp -P "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/${OUTPUT_FILE}" "${ROOT_DIR}/${OUTPUT_FILE}"

echo "[4/4] Done! Export saved to ${ROOT_DIR}/${OUTPUT_FILE}"
