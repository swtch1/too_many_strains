#!/usr/bin/env bash

# this script wraps the database migration script so that we can both, ensure
# the script runs with the correct environment, and meets the requirement that
# the database creation/migration script "run by itself using `go run`"
# while still utilizing the CLI it deserves.

set -euo pipefail

SCRIPT_DIR="cmd/database-migration"
SEED_FILE="strains.json"

function cleanup() {
  # if we make it to the script dir, cleanup the config from there
  if [[ $(basename "$(pwd)") == "database-migration" ]];then
    rm -f "$SEED_FILE"
  else
    rm -f "${SCRIPT_DIR}/${SEED_FILE}"
  fi
}

# ensure cleanup on exit
trap cleanup EXIT

# make sure we're in the right directory
cd "$(dirname "$0")"

cp "$SEED_FILE" "$SCRIPT_DIR"

cd "$SCRIPT_DIR"

go run -mod=vendor .
