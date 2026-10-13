#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/fuel_consumptions-docker.sh
"$FUEL_CONSUMPTIONS_DOCKER" compose --profile checks run --rm --build fuel_consumptions-checks
