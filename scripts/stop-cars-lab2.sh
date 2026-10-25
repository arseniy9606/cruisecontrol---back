#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/cars-docker.sh
"$CARS_DOCKER" compose stop
echo "ЛР2 остановлена. Все данные сохранены."
