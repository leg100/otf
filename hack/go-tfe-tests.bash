#!/usr/bin/env bash

# Run go-tfe's tests against an otfd instance. Either specify tests as arguments
# or a default subset of tests will be run.

set -ex

export TFE_TOKEN=3728a09aec6a714bcd57498865072e042f5a63347a930420707edc8c213843c6
export TFE_ADDRESS=https://localhost:8080
export SKIP_PAID=1

# Default subset of tests
TESTS="${@:-Test(Workspaces(Create|List|Update|Delete|Unlock|Lock)|Organizations(Create|List|Read|Update)|StateVersions|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

cd $(GOPROXY=direct go mod download -json github.com/leg100/go-tfe@otf | jq -r '.Dir')
go test -v -run $TESTS -timeout 60s
