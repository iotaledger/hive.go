#!/usr/bin/env bash
set -euxo pipefail

go version
go clean -testcache

find . -name go.mod -print0 |  xargs -0 -n1 dirname | xargs -t -n1 -I {} bash -c 'set -euxo pipefail && cd "{}" && go test -tags rocksdb,stacktrace ./...'