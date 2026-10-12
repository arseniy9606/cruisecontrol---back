#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if command -v go >/dev/null 2>&1; then
  FUEL_CONSUMPTIONS_GO=$(command -v go)
elif [ -x .tools/go/bin/go ]; then
  FUEL_CONSUMPTIONS_GO="$PWD/.tools/go/bin/go"
elif [ -x ../../work/tools/go/bin/go ]; then
  FUEL_CONSUMPTIONS_GO="$PWD/../../work/tools/go/bin/go"
else
  echo "Go не найден. Запустите scripts/install-fuel_consumptions-go.sh."
  exit 1
fi
mkdir -p .runtime/go-cache .runtime/go-path
export GOCACHE="$PWD/.runtime/go-cache" GOPATH="$PWD/.runtime/go-path" CGO_ENABLED=0
"$FUEL_CONSUMPTIONS_GO" test -buildvcs=false ./...
"$FUEL_CONSUMPTIONS_GO" vet -buildvcs=false ./...
