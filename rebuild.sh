#!/usr/bin/env bash

set -e

./docker_compose.sh build --parallel "${@}"
./docker_compose.sh stop "${@}"
./docker_compose.sh rm -f "${@}"
./docker_compose.sh up -d "${@}"
