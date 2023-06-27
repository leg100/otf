#!/usr/bin/env bash

# Run go-tfe tests against otfd. It's recommended you first start otfd and
# postgres using docker compose before running this script. This script
# expects to find otfd running on port 8833 which is the port the docker
# compose otfd listens on.
#
# Either specify tests as arguments or a default subset of tests will be run.

set -e

GO_TFE_REPO="${GO_TFE_REPO:-github.com/leg100/go-tfe@otf}"
TESTS="${@:-Test(Variables|Workspaces(Create|List|Update|Delete|Lock|Unlock|ForceUnlock|Read\$|ReadByID)|Organizations(Create|List|Read|Update)|StateVersion|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

export TFE_ADDRESS="${TFE_ADDRESS:-https://localhost:8080}"
# go-tfe tests perform privileged operations (e.g. creating organizations), so
# we use a site admin token.
#
# NOTE: this token is the same token specified in docker compose
export TFE_TOKEN=${TFE_TOKEN:-site-token}
export SKIP_PAID=1
export SSL_CERT_FILE=$PWD/internal/integration/fixtures/cert.pem

cd $(go mod download -json ${GO_TFE_REPO} | jq -r '.Dir')
go test -v -run $TESTS -timeout 60s
