#!/usr/bin/env bash

set -e

function cleanup() {
  ./docker-compose.sh down --remove-orphans --volumes
}
trap cleanup exit

./docker-compose.sh up "${@}"
