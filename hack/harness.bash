#!/usr/bin/env bash

# A test harness for otsd. It'll first check if otsd is running. If it
# is not running it'll start otsd before the tests, and terminate
# it afterwards (Often a developer instead runs otsd in another
# terminal...).


set -x

export PATH=$PWD/_build:$PATH

export OTS_SSL=true
export OTS_CERT_FILE=./e2e/fixtures/cert.crt
export OTS_KEY_FILE=./e2e/fixtures/key.pem
export OTS_DB_PATH=$(mktemp)

# Track whether this script started otsd
started=0

# Upon exit, stop otsd if this script started it
function cleanup()
{
    if [[ $started -eq 1 ]]; then
        pkill otsd
    fi
}
trap cleanup EXIT

# Upon error, print out otsd logs (...but only if this script started it),
# and exit
function print_logs()
{
    if [[ $started -eq 1 ]]; then
        echo "--- otsd output ---"
        echo
        cat otsd.log
    fi
}
trap print_logs ERR

# Start otsd if not already running
if ! pgrep otsd; then
    nohup otsd > otsd.log &
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
