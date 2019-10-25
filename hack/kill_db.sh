#!/usr/bin/env bash

set -euo pipefail

docker container kill flourish-mysql &>/dev/null
