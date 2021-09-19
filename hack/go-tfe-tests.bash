#!/usr/bin/env bash

# Run go-tfe's tests against an otfd instance. Either specify tests as arguments
# or a default subset of tests will be run.

set -ex

export TFE_TOKEN=dummy
export TFE_ADDRESS=https://localhost:8080
export SKIP_PAID=1

TESTS="${@:-Test(Workspaces(Create|List|Update|Delete|Unlock|Lock)|Organizations(Create|List|Read|Update)|StateVersions|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

cd $(go list -f '{{.Dir}}' github.com/leg100/go-tfe)
go test -v -run $TESTS
