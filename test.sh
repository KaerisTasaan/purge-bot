#!/usr/bin/env sh
# Run tests the same way as CI (pure Go SQLite; no C compiler required).
set -e
go mod download
go mod verify
go test -v ./... -count=1
