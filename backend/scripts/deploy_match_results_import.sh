#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

SERVER_HOST="${SERVER_HOST:-203.0.113.10}"
SERVER_USER="${SERVER_USER:-root}"
SERVER_PORT="${SERVER_PORT:-22}"
REMOTE_DIR="${REMOTE_DIR:-/opt/soulink-match-import}"
REMOTE_CONFIG_DIR="${REMOTE_DIR}/configs/match"
POSTGRES_CONTAINER="${POSTGRES_CONTAINER:-soulink-postgres}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
OUTPUT_BINARY="${OUTPUT_BINARY:-matchresultsimport}"
IMPORT_CONFIG_FILE="${IMPORT_CONFIG_FILE:-configs/match/match_results_import.local.json}"
RESULTS_CSV_FILE="${RESULTS_CSV_FILE:-results/match_results_r5.csv}"
EXPLANATIONS_CSV_FILE="${EXPLANATIONS_CSV_FILE:-results/match_explanations_r5.csv}"

if [[ -z "${SERVER_HOST}" ]]; then
  echo "SERVER_HOST is required, for example: SERVER_HOST=203.0.113.10 $0" >&2
  exit 1
fi

cd "${ROOT_DIR}"

echo "[1/5] Building matchresultsimport for linux/${GOARCH_TARGET}..."
CGO_ENABLED=0 GOOS=linux GOARCH="${GOARCH_TARGET}" go build -trimpath -o "${OUTPUT_BINARY}" ./cmd/matchresultsimport

echo "[2/5] Preparing remote directories on ${SERVER_USER}@${SERVER_HOST}..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "mkdir -p '${REMOTE_CONFIG_DIR}'"

echo "[3/5] Uploading binary, config and CSV..."
scp -P "${SERVER_PORT}" "${OUTPUT_BINARY}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/"
scp -P "${SERVER_PORT}" "${ROOT_DIR}/${IMPORT_CONFIG_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_CONFIG_DIR}/match_results_import.local.json"
scp -P "${SERVER_PORT}" "${RESULTS_CSV_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/match_results.csv"
IMPORT_EXPLANATIONS_ARG=""
if [[ -n "${EXPLANATIONS_CSV_FILE}" && -f "${EXPLANATIONS_CSV_FILE}" ]]; then
  scp -P "${SERVER_PORT}" "${EXPLANATIONS_CSV_FILE}" "${SERVER_USER}@${SERVER_HOST}:${REMOTE_DIR}/match_explanations.csv"
  IMPORT_EXPLANATIONS_ARG="-explanations './match_explanations.csv'"
else
  echo "Explanation CSV not found, skip uploading: ${EXPLANATIONS_CSV_FILE}"
fi
echo "[4/5] Running import on remote server..."
ssh -p "${SERVER_PORT}" "${SERVER_USER}@${SERVER_HOST}" "\
cd '${REMOTE_DIR}' && \
chmod +x './${OUTPUT_BINARY}' && \
rm -f .env && \
PG_IP=\$(docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' '${POSTGRES_CONTAINER}') && \
DB_PASS=\$(docker exec '${POSTGRES_CONTAINER}' printenv POSTGRES_PASSWORD) && \
DB_DSN=\"postgres://solink:\${DB_PASS}@\${PG_IP}:5432/soulink-web?sslmode=disable\" \
DB_AUTO_MIGRATE=false \
'./${OUTPUT_BINARY}' -config '${REMOTE_CONFIG_DIR}/match_results_import.local.json' -csv './match_results.csv' ${IMPORT_EXPLANATIONS_ARG}"

echo "[5/5] Cleaning local binary..."
rm -f "${OUTPUT_BINARY}"
echo "Done! Match results have been imported to the remote database."
