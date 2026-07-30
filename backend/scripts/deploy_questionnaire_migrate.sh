#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVER_HOST="${SERVER_HOST:-}"
SERVER_USER="${SERVER_USER:-root}"
SERVER_PORT="${SERVER_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/opt/soulink-migrate}"
REMOTE_CONFIG_DIR="${REMOTE_DIR}/configs/questionnaire"
REMOTE_POSTGRES_CONTAINER="${REMOTE_POSTGRES_CONTAINER:-soulink-postgres}"
POSTGRES_DB="${POSTGRES_DB:-soulink-web}"
POSTGRES_USER="${POSTGRES_USER:-solink}"
POSTGRES_PASSWORD_ENV="${POSTGRES_PASSWORD_ENV:-POSTGRES_PASSWORD}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"

SOURCE_CATALOG_FILE="${SOURCE_CATALOG_FILE:-}"
TARGET_CATALOG_FILE="${TARGET_CATALOG_FILE:-}"
MIGRATION_CONFIG_FILE="${MIGRATION_CONFIG_FILE:-configs/questionnaire/answer_migration.local.json}"
OUTPUT_BINARY="${OUTPUT_BINARY:-questionnairemigrate}"

if [[ -z "${SERVER_HOST}" ]]; then
  echo "SERVER_HOST is required, for example: SERVER_HOST=203.0.113.10 $0" >&2
  exit 1
fi
if [[ -z "${SOURCE_CATALOG_FILE}" ]]; then
  echo "SOURCE_CATALOG_FILE is required, for example: SOURCE_CATALOG_FILE=thu_questionnaire_template_v0.1.md" >&2
  exit 1
fi
if [[ -z "${TARGET_CATALOG_FILE}" ]]; then
  echo "TARGET_CATALOG_FILE is required, for example: TARGET_CATALOG_FILE=thu_questionnaire_template.md" >&2
  exit 1
fi

cd "${ROOT_DIR}"

echo "[1/5] Building questionnairemigrate for linux/${GOARCH_TARGET}..."
CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH_TARGET}" go build -trimpath -o "${OUTPUT_BINARY}" ./cmd/questionnairemigrate

echo "[2/5] Preparing remote directories on ${SERVER_USER}@${SERVER_HOST}..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p '${REMOTE_CONFIG_DIR}'"

echo "[3/5] Uploading binary and migration files..."
scp -P "${SERVER_PORT}" "${OUTPUT_BINARY}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/${MIGRATION_CONFIG_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CONFIG_DIR}/answer_migration.local.json"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/configs/questionnaire/${SOURCE_CATALOG_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CONFIG_DIR}/"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/configs/questionnaire/${TARGET_CATALOG_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CONFIG_DIR}/"

echo "[4/5] Repairing legacy questionnaire constraints on remote database..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "chmod +x '${REMOTE_DIR}/${OUTPUT_BINARY}' && docker exec -i '${REMOTE_POSTGRES_CONTAINER}' psql -U '${POSTGRES_USER}' -d '${POSTGRES_DB}' <<'SQL'
ALTER TABLE user_questionnaire_answers
    DROP CONSTRAINT IF EXISTS user_questionnaire_answers_user_id_questionnaire_key_questi_key;

ALTER TABLE user_questionnaire_answers
    DROP CONSTRAINT IF EXISTS user_questionnaire_answers_user_id_questionnaire_key_question_id_key;

DO \$\$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conrelid = 'user_questionnaire_answers'::regclass
          AND conname = 'user_questionnaire_answers_user_key_version_question_key'
    ) THEN
        ALTER TABLE user_questionnaire_answers
            ADD CONSTRAINT user_questionnaire_answers_user_key_version_question_key
            UNIQUE (user_id, questionnaire_key, questionnaire_version, question_id);
    END IF;
END
\$\$;
SQL"

echo "[5/5] Running questionnaire migration on remote host via Docker network..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "\
cd '${REMOTE_DIR}' && \
chmod +x './${OUTPUT_BINARY}' && \
rm -f .env && \
PG_IP=\$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' '${REMOTE_POSTGRES_CONTAINER}') && \
DB_PASS=\$(docker exec '${REMOTE_POSTGRES_CONTAINER}' printenv '${POSTGRES_PASSWORD_ENV}') && \
DB_DSN=\"postgres://${POSTGRES_USER}:\${DB_PASS}@\${PG_IP}:5432/${POSTGRES_DB}?sslmode=disable\" \
DB_AUTO_MIGRATE=false \
'./${OUTPUT_BINARY}' -config './configs/questionnaire/answer_migration.local.json'"

rm -f "${OUTPUT_BINARY}"
echo "Done! Questionnaire migration completed through the Docker internal network without exposing database ports."