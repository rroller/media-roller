#!/usr/bin/env bash
set -ex

go get -u ./...
go mod tidy
go fmt ./...

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.64.5
$(go env GOPATH)/bin/golangci-lint --version
$(go env GOPATH)/bin/golangci-lint run
