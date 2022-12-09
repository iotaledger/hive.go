#!/bin/bash
set -euxo pipefail

FILTER="$1"

function lint {
    if [ "$1" == "" ]; then
        golangci-lint run
    else
        golangci-lint run | grep --color "$1"
    fi
}

for p in apputils core serializer
do
    echo
    echo "entering $p..."
    echo
    cd $p
    lint "$FILTER"
    cd ..
done
