#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/fuel_consumptions-docker.sh
"$FUEL_CONSUMPTIONS_DOCKER" compose stop
echo "ЛР2 остановлена. Все данные сохранены."
