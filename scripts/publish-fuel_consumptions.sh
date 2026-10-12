#!/bin/sh
set -eu
cd "$(dirname "$0")/.."
git push -u https://github.com/arseniy9606/cruisecontrol---back.git HEAD:refs/heads/lab1-fuel_consumptions
