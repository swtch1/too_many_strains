#!/usr/bin/env bash

set -euo pipefail

docker run --name flourish-mysql --publish 3306:3306 -d --rm flourish-mysql:latest
