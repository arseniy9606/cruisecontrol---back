#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if command -v docker >/dev/null 2>&1; then
 FUEL_CONSUMPTIONS_DOCKER=$(command -v docker)
elif [ -x /Applications/Docker.app/Contents/Resources/bin/docker ]; then
 FUEL_CONSUMPTIONS_DOCKER=/Applications/Docker.app/Contents/Resources/bin/docker
else
 echo "Установите и запустите Docker Desktop. Установщик: https://www.docker.com/products/docker-desktop/"
 exit 1
fi
if ! "$FUEL_CONSUMPTIONS_DOCKER" info >/dev/null 2>&1; then
 echo "Запустите Docker Desktop и дождитесь Engine running, затем повторите запуск."
 exit 1
fi
"$FUEL_CONSUMPTIONS_DOCKER" compose up -d --build fuel_consumptions-minio
"$FUEL_CONSUMPTIONS_DOCKER" compose run --build --rm fuel_consumptions-media-init
printf 'MinIO: http://localhost:9001 · root / rootpassword\n'
printf 'Медиа: http://localhost:9000/fuel-consumptions/\n'
