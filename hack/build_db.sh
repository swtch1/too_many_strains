#!/usr/bin/env bash

set -euo pipefail

cd "$(dirname "$0")"  && cd ../reference/Docker

docker build . -t flourish-mysql
