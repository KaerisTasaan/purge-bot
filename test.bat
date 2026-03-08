@echo off
REM Run tests the same way as CI (pure Go SQLite; no C compiler required).
go mod download
go mod verify
go test -v ./... -count=1
