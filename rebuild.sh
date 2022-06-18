#!/usr/bin/env bash

set -e

./docker-compose.sh build --parallel "${@}"
./docker-compose.sh stop "${@}"
./docker-compose.sh rm -f "${@}"
./docker-compose.sh up -d "${@}"
