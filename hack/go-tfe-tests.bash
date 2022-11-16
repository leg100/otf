#!/usr/bin/env bash

# Run go-tfe's tests against an otfd instance. Either specify tests as arguments
# or a default subset of tests will be run.

set -e

# go-tfe tests perform privileged operations (e.g.
# creating organizations), so we use a site admin token.
SITE_TOKEN=go-tfe-test-site-token

# run otfd on random port in background, logging to a temp file
logfile=$(tempfile)
nohup _build/otfd --address :0 --log-level trace --log-http-requests --site-token $SITE_TOKEN > $logfile 2>&1 &
pid=$!

# print out logs upon error
function print_logs()
{
    echo "--- otfd output ---"
    echo
    cat $logfile
}
trap print_logs ERR

# stop otfd upon exit
function cleanup()
{
    kill $pid
    # wait til it's dead
    while [[ -d /proc/$pid ]]; do
        sleep 1
    done
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
export OTF_SSL=true
export OTF_CERT_FILE=./e2e/fixtures/cert.crt
export OTF_KEY_FILE=./e2e/fixtures/key.pem

TESTS="${@:-Test(Workspaces(Create|List|Update|Delete|Unlock|Lock|Read\$|ReadByID)|Organizations(Create|List|Read|Update)|StateVersions|Runs|Plans|Applies(Read|Logs)|ConfigurationVersions)}"

cd $(GOPROXY=direct go mod download -json github.com/leg100/go-tfe@otf | jq -r '.Dir')
go test -v -run $TESTS -timeout 60s
