#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/fuel_consumptions-docker.sh
[ -f .env ] || cp .env.example .env
"$FUEL_CONSUMPTIONS_DOCKER" compose up -d --build
FUEL_CONSUMPTIONS_ATTEMPT=0
until curl -fsS http://localhost:8082/fuel_consumptions >/dev/null 2>&1; do
  FUEL_CONSUMPTIONS_ATTEMPT=$((FUEL_CONSUMPTIONS_ATTEMPT + 1))
  [ "$FUEL_CONSUMPTIONS_ATTEMPT" -lt 60 ] || { "$FUEL_CONSUMPTIONS_DOCKER" compose logs --tail=40 fuel_consumptions-backend; exit 1; }
  sleep 2
done
echo "ЛР2 работает: http://localhost:8082/fuel_consumptions"
echo "Adminer: http://localhost:8083 · MinIO: http://localhost:9101"
