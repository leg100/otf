#!/usr/bin/env bash

# Run go-tfe's tests against an otfd instance. Either specify tests as arguments
# or a default subset of tests will be run.

set -e

if [ -z $OTF_TEST_DATABASE_URL ]; then
    echo OTF_TEST_DATABASE_URL not set
    exit 1
fi

export OTF_SSL=true
export OTF_CERT_FILE=./integration/cert.pem
export OTF_KEY_FILE=./integration/key.pem
export SSL_CERT_FILE=$PWD/integration/fixtures/cert.pem

# go-tfe tests perform privileged operations (e.g.
# creating organizations), so we use a site admin token.
SITE_TOKEN=go-tfe-test-site-token

# run otfd on random port in background, logging to a temp file
logfile=$(mktemp)
nohup _build/otfd --address :0 \
    --log-level trace --log-http-requests \
    --site-token $SITE_TOKEN \
    --secret ce6bf87f25118c87c8ca3d3066010c5ee56643c01ba5cab605642b0d83271e6e \
    --ssl true \
    --dev-mode=false \
    --cert-file ./integration/fixtures/cert.pem \
    --key-file ./integration/fixtures/key.pem \
    --database $OTF_TEST_DATABASE_URL > $logfile 2>&1 &
pid=$!

# stop otfd upon exit
function cleanup()
{
    otfd_ecode=$?
    [[ -d /proc/$pid ]] && kill $pid
    # wait til it's dead
    while [[ -d /proc/$pid ]]; do
        sleep 1
    done
    # print out logs upon error
    if [ "${otfd_ecode}" != "0" ]; then
        echo "--- otfd output ---"
        echo
        cat $logfile
    fi
}
trap cleanup EXIT

# wait for otfd to listen on port and capture port number
tries=0
while true; do
    port=$(lsof -w -p $pid -a -s TCP:LISTEN -i TCP -Fn | awk -F: '$1 == "n*" { print $2 }')
    [[ -n $port ]] && break

    tries=$((tries+1))
    if [[ tries -gt 5 ]]; then
        echo failed to find listening port for otfd
        exit 1
    fi
    if [[ ! -d /proc/$pid ]]; then
        echo otfd died prematurely
        exit 1
    fi
done

export TFE_ADDRESS="https://localhost:${port}"
export TFE_TOKEN=$SITE_TOKEN
export SKIP_PAID=1

TESTS="${@:-Test(Variables|Workspaces(Create|List|Update|Delete|Unlock|Lock|Read\$|ReadByID)|Organizations(Create|List|Read|Update)|StateVersion|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

cd $(go mod download -json github.com/leg100/go-tfe@otf | jq -r '.Dir')
go test -v -run $TESTS -timeout 60s
