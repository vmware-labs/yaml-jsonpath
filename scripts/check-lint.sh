#!/usr/bin/env bash
#
# run golangci-lint

golangci-lint -E bodyclose,misspell,gofmt,golint,unconvert,goimports,depguard,gocritic,interfacer run
