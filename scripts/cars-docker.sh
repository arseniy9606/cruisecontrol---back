#!/bin/sh
set -eu
if command -v docker >/dev/null 2>&1; then
  CARS_DOCKER=$(command -v docker)
elif [ -x /Applications/Docker.app/Contents/Resources/bin/docker ]; then
  CARS_DOCKER=/Applications/Docker.app/Contents/Resources/bin/docker
else
  echo "Установите Docker Desktop: https://www.docker.com/products/docker-desktop/"; exit 1
fi
export PATH="$(dirname "$CARS_DOCKER"):$PATH"
if ! "$CARS_DOCKER" info >/dev/null 2>&1; then
  if [ "$(uname)" = Darwin ]; then open -a Docker; fi
  echo "Ожидание запуска Docker Desktop..."
  CARS_ATTEMPT=0
  until "$CARS_DOCKER" info >/dev/null 2>&1; do
    CARS_ATTEMPT=$((CARS_ATTEMPT + 1))
    [ "$CARS_ATTEMPT" -lt 60 ] || { echo "Откройте Docker Desktop и повторите запуск."; exit 1; }
    sleep 2
  done
fi
