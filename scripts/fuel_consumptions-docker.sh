#!/bin/sh
set -eu
if command -v docker >/dev/null 2>&1; then
  FUEL_CONSUMPTIONS_DOCKER=$(command -v docker)
elif [ -x /Applications/Docker.app/Contents/Resources/bin/docker ]; then
  FUEL_CONSUMPTIONS_DOCKER=/Applications/Docker.app/Contents/Resources/bin/docker
else
  echo "Установите Docker Desktop: https://www.docker.com/products/docker-desktop/"; exit 1
fi
export PATH="$(dirname "$FUEL_CONSUMPTIONS_DOCKER"):$PATH"
if ! "$FUEL_CONSUMPTIONS_DOCKER" info >/dev/null 2>&1; then
  if [ "$(uname)" = Darwin ]; then open -a Docker; fi
  echo "Ожидание запуска Docker Desktop..."
  FUEL_CONSUMPTIONS_ATTEMPT=0
  until "$FUEL_CONSUMPTIONS_DOCKER" info >/dev/null 2>&1; do
    FUEL_CONSUMPTIONS_ATTEMPT=$((FUEL_CONSUMPTIONS_ATTEMPT + 1))
    [ "$FUEL_CONSUMPTIONS_ATTEMPT" -lt 60 ] || { echo "Откройте Docker Desktop и повторите запуск."; exit 1; }
    sleep 2
  done
fi
