#!/usr/bin/env sh
set -eu

CONTAINER_NAME="${1:-pg-notify-db}"
BATCH_SIZE="${2:-50}"
SLEEP_SECONDS="${3:-1}"

while true; do
  docker exec -i "$CONTAINER_NAME" psql -U postgres -d pg_notify -c "SELECT simulate_activity(${BATCH_SIZE});" >/dev/null
  sleep "$SLEEP_SECONDS"
done
