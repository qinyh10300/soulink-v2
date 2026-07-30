#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVER_HOST="${SERVER_HOST:-203.0.113.10}"
SERVER_USER="${SERVER_USER:-root}"
SERVER_PORT="${SERVER_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/opt/soulink-export}"
REMOTE_CATALOG_DIR="${REMOTE_DIR}/configs/questionnaire"
CATALOG_FILE="${CATALOG_FILE:-thu_questionnaire_template.md}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
OUTPUT_BINARY="${OUTPUT_BINARY:-questionnaireexport}"
COMPLETED_ONLY="${COMPLETED_ONLY:-false}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-soulink-postgres}"

cd "${ROOT_DIR}"

echo "[1/5] Building questionnaireexport for linux/${GOARCH_TARGET}..."
CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH_TARGET}" go build -trimpath -o "${OUTPUT_BINARY}" ./cmd/questionnaireexport

echo "[2/5] Preparing remote directory on ${SERVER_USER}@${SERVER_HOST}..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p '${REMOTE_CATALOG_DIR}'"

echo "[3/5] Uploading binary and config files..."
scp -P "${SERVER_PORT}" "${OUTPUT_BINARY}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/configs/questionnaire/${CATALOG_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CATALOG_DIR}/"

EXTRA_FLAGS=""
if [[ "${COMPLETED_ONLY}" == "true" ]]; then
  EXTRA_FLAGS="-completed-only"
fi

echo "[4/5] Running export on remote server..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "\
cd '${REMOTE_DIR}' && \
chmod +x './${OUTPUT_BINARY}' && \
rm -f .env && \
PG_IP=\$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' '${POSTGRES_CONTAINER}') && \
DB_PASS=\$(docker exec '${POSTGRES_CONTAINER}' printenv POSTGRES_PASSWORD) && \
DB_DSN=\"postgres://solink:\${DB_PASS}@\${PG_IP}:5432/soulink-web?sslmode=disable\" \
DB_AUTO_MIGRATE=false \
'./${OUTPUT_BINARY}' -catalog '${REMOTE_CATALOG_DIR}/${CATALOG_FILE}' -output export.csv ${EXTRA_FLAGS}"

echo "[5/5] Downloading export CSV..."
REMOTE_CSV=$(ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "ls -t ${REMOTE_DIR}/export_*.csv 2>/dev/null | head -1")
if [[ -z "${REMOTE_CSV}" ]]; then
  echo "ERROR: No versioned export CSV found on remote server." >&2
  exit 1
fi
LOCAL_CSV="${ROOT_DIR}/$(basename "${REMOTE_CSV}")"
scp -P "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CSV}" "${LOCAL_CSV}"

rm -f "${OUTPUT_BINARY}"
echo "Done! Export saved to ${LOCAL_CSV}"
