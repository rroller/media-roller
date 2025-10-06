#!/usr/bin/env bash
set -ex

go get -u ./...
go mod tidy
go fmt ./...

golangci-lint --version
golangci-lint run
