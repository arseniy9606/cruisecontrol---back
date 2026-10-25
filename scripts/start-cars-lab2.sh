#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
. ./scripts/cars-docker.sh
[ -f .env ] || cp .env.example .env
"$CARS_DOCKER" compose up -d --build
CARS_ATTEMPT=0
until curl -fsS http://localhost:8082/cars >/dev/null 2>&1; do
  CARS_ATTEMPT=$((CARS_ATTEMPT + 1))
  [ "$CARS_ATTEMPT" -lt 60 ] || { "$CARS_DOCKER" compose logs --tail=40 cars-backend; exit 1; }
  sleep 2
done
echo "ЛР2 работает: http://localhost:8082/cars"
echo "Adminer: http://localhost:8083 · MinIO: http://localhost:9101"
