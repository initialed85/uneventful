#!/usr/bin/env bash

set -e

DOCKER_COMPOSE="docker compose -f docker/docker-compose.yml"

${DOCKER_COMPOSE} "${@}"
