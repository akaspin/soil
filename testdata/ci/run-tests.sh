#!/usr/bin/env bash

docker images
#docker pull fedora

set -e

echo "tests with race detection"
go test -tags="test_unit test_cluster test_systemd" -p=1 -race ./...
