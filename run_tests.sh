#!/usr/bin/env bash
set -euxo pipefail

find . -name go.mod -execdir go test ./... \;