#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/cars-docker.sh
"$CARS_DOCKER" compose --profile checks run --rm --build cars-checks
