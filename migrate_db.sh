#!/usr/bin/env bash

# this script wraps the database migration script so that we can both, ensure
# the script runs with the correct environment, and meets the requirement that
# the database creation/migration script "run by itself using `go run`"
# while still utilizing the CLI it deserves.

set -euo pipefail

SCRIPT_DIR="cmd/database-migration"
CONFIG_FILE="config.yaml"

function cleanup() {
  # if we make it to the script dir, cleanup the config from there
  if [[ $(basename "$(pwd)") == "database-migration" ]];then
    rm -f "$CONFIG_FILE"
  else
    rm -f "${SCRIPT_DIR}/${CONFIG_FILE}"
  fi
}

# ensure cleanup on exit
trap cleanup EXIT

# make sure we're in the right directory
cd "$(dirname "$0")"

cp "$CONFIG_FILE" "$SCRIPT_DIR"

cd "$SCRIPT_DIR"

go run . -mod=vendor

rm -f "$CONFIG_FILE"
