#!/usr/bin/env bash

# A test harness for otfd. It'll first check if otfd is running. If it
# is not running it'll start otfd before the tests, and terminate
# it afterwards (Often a developer instead runs otfd in another
# terminal...).


set -x

export PATH=$PWD/_build:$PATH

export OTF_SSL=true
export OTF_CERT_FILE=./e2e/fixtures/cert.crt
export OTF_KEY_FILE=./e2e/fixtures/key.pem
export OTF_SITE_TOKEN=be68d73c52054e21685a53c483275e054f05c7844907c103ca2792d7cf9fe642

# Track whether this script started otfd
started=0

# Upon exit, stop otfd if this script started it
function cleanup()
{
    if [[ $started -eq 1 ]]; then
        pkill otfd
        # wait til it's dead
        while pgrep otfd; do
            sleep 1
        done
    fi
}
trap cleanup EXIT

# Upon error, print out otfd logs (...but only if this script started it),
# and exit
function print_logs()
{
    if [[ $started -eq 1 ]]; then
        echo "--- otfd output ---"
        echo
        cat otfd.log
    fi
}
trap print_logs ERR

# Start otfd if not already running
if ! pgrep otfd; then
    nohup otfd --log-level trace --log-color true > otfd.log &
    started=1
fi

# Wait til it's running
curl \
    --retry 5 \
    --retry-connrefused \
    -H'Accept: application/vnd.api+json' \
    https://localhost:8080/api/v2/ping

# Run tests...
$@
