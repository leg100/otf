#!/usr/bin/env bash

# Run go-tfe's tests against an otfd instance. Either specify tests as arguments
# or a default subset of tests will be run.

set -ex

export TFE_TOKEN=dummy
export TFE_ADDRESS=https://localhost:8080
export SKIP_PAID=1

# Default subset of tests
TESTS="${@:-Test(Workspaces(Create|List|Update|Delete|Unlock|Lock)|Organizations(Create|List|Read|Update)|StateVersions|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

cd $(go mod download -json github.com/hashicorp/go-tfe@v0.17.1 | jq -r '.Dir')
go test -v -run $TESTS
