#!/usr/bin/env bash

set -ex

export TFE_TOKEN=dummy
export TFE_ADDRESS=https://localhost:8080

cd $(go list -m -u -json github.com/leg100/go-tfe | jq -r '.Dir')
go test -v -run TestWorkspacesList
