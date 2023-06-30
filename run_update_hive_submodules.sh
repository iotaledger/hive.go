#!/bin/bash
COMMIT=$1

if [ -z "$COMMIT" ]; then
    echo "ERROR: no commit hash given!"
    exit 1
fi

SUBMODULES=$(find . -name "go.mod" -exec dirname {} \; | sed -e 's/^\.\///' | sort)

for submodule in $SUBMODULES; do
    pushd "$submodule" >/dev/null
    echo "updating ${submodule}..."
    hivemodules=$(grep '^\sgithub.com/iotaledger/hive.go' go.mod | awk '{print $1}')
    for hivemodule in $hivemodules; do
        echo "   go get -u ${hivemodule}..."
        go get -u "$hivemodule@$COMMIT"
    done
    go mod tidy >/dev/null

    popd >/dev/null
done
