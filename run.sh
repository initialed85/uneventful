#!/usr/bin/env bash

set -e

function cleanup() {
  ./docker_compose.sh down --remove-orphans --volumes
}
trap cleanup exit

./docker_compose.sh up "${@}"
