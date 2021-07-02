#!/usr/bin/env bash

set -ex

export TFE_TOKEN=dummy
export TFE_ADDRESS=https://localhost:8080

cd $(go list -f '{{.Dir}}' github.com/leg100/go-tfe)
go test -v -run 'TestWorkspaces(Create|List|Update|Unlock|Lock)'
