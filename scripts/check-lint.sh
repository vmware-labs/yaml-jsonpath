#!/usr/bin/env bash
#
# run golangci-lint

golangci-lint -E bodyclose,misspell,dupl,gofmt,golint,unconvert,goimports,depguard,gocritic,interfacer run
