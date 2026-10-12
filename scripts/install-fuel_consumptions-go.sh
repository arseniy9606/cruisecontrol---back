#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
if [ "$(uname -s)" != Darwin ] || [ "$(uname -m)" != arm64 ]; then
 echo "Этот установщик подготовлен для macOS Apple Silicon. Для другой системы установите Go с https://go.dev/dl/."
 exit 1
fi
mkdir -p .tools
curl --fail --location https://go.dev/dl/go1.27.2.darwin-arm64.tar.gz --output .tools/fuel_consumptions-go.tar.gz
printf '76812b213b1b2302c978d28fa52fa92d541704b9e7d9d5db8002c50e4018c4c5  .tools/fuel_consumptions-go.tar.gz\n' | shasum -a 256 -c -
tar -xzf .tools/fuel_consumptions-go.tar.gz -C .tools
.tools/go/bin/go version
