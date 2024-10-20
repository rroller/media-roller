#!/usr/bin/env bash
set -ex

go mod tidy
go fmt ./...