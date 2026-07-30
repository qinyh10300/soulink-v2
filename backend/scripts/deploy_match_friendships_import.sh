#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# 203.0.113.10
# 203.0.113.10
SERVER_HOST="${SERVER_HOST:-203.0.113.10}"
SERVER_USER="${SERVER_USER:-root}"
SERVER_PORT="${SERVER_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/opt/soulink-match-friends}"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-soulink-postgres}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
OUTPUT_BINARY="${OUTPUT_BINARY:-matchfriendsimport}"
RESULTS_CSV_FILE="${RESULTS_CSV_FILE:-results/match_results_r5.csv}"
DRY_RUN="${DRY_RUN:-false}"
ENSURE_CONVERSATIONS="${ENSURE_CONVERSATIONS:-false}"
# 设为 true：只构建并上传二进制与 CSV，不在远端执行导入（等价于部署完即返回）
UPLOAD_ONLY="${UPLOAD_ONLY:-false}"

if [[ -z "${SERVER_HOST}" ]]; then
  echo "SERVER_HOST is required, for example: SERVER_HOST=203.0.113.10 $0" >&2
  exit 1
fi

if [[ ! -f "${ROOT_DIR}/${RESULTS_CSV_FILE}" ]]; then
  echo "RESULTS_CSV_FILE not found: ${ROOT_DIR}/${RESULTS_CSV_FILE}" >&2
  exit 1
fi

cd "${ROOT_DIR}"

REMOTE_IMPORT_ARGS=""
if [[ "${DRY_RUN}" == "true" ]]; then
  REMOTE_IMPORT_ARGS="${REMOTE_IMPORT_ARGS} -dry-run"
fi
if [[ "${ENSURE_CONVERSATIONS}" == "true" ]]; then
  REMOTE_IMPORT_ARGS="${REMOTE_IMPORT_ARGS} -ensure-conversations"
fi

echo "[1/4] Building matchfriendsimport for linux/${GOARCH_TARGET}..."
CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH_TARGET}" go build -trimpath -o "${OUTPUT_BINARY}" ./cmd/matchfriendsimport

echo "[2/4] Preparing remote directory on ${SERVER_USER}@${SERVER_HOST}..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p '${REMOTE_DIR}'"

echo "[3/4] Uploading binary and CSV..."
scp -P "${SERVER_PORT}" "${OUTPUT_BINARY}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/${RESULTS_CSV_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/match_results.csv"

if [[ "${UPLOAD_ONLY}" == "true" ]]; then
  echo "[4/4] Skipped remote run (UPLOAD_ONLY=true). Files on ${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"
else
  echo "[4/4] Running friend import on remote server..."
  ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "\
cd '${REMOTE_DIR}' && \
chmod +x './${OUTPUT_BINARY}' && \
rm -f .env && \
PG_IP=\$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' '${POSTGRES_CONTAINER}') && \
DB_PASS=\$(docker exec '${POSTGRES_CONTAINER}' printenv POSTGRES_PASSWORD) && \
DB_DSN=\"postgres://solink:\${DB_PASS}@\${PG_IP}:5432/soulink-web?sslmode=disable\" \
DB_AUTO_MIGRATE=false \
'./${OUTPUT_BINARY}' -csv './match_results.csv'${REMOTE_IMPORT_ARGS}"
fi

echo "Cleaning local binary..."
rm -f "${OUTPUT_BINARY}"
if [[ "${UPLOAD_ONLY}" == "true" ]]; then
  echo "Done! Upload complete; remote import was not run."
else
  echo "Done! Match friendships import has finished."
fi
