#!/usr/bin/env bash

#docker images
#docker pull fedora

set -e

echo "coverage"
echo "" > coverage.txt

for d in $(go list ./... | grep -v vendor); do
    go test -tags="test_unit test_cluster test_systemd" -p=1 -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
