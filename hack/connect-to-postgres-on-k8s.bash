#!/usr/bin/env bash

# Connect to postgres on kubernetes cluster. Credit goes to @ringerc for code
# to capture the random local port from `kubectl port-forward`:
#
# https://github.com/kubernetes/kubectl/issues/1190#issuecomment-1075911615

set -e

PORT_FORWARD_TIMEOUT_SECONDS=10

# Start a kubectl port-forward and wait until the port is active
#
# The local port is automatically selected and is returned as global variable
#
#    port_forward_local_port
#
function start_port_forward() {
    coproc kubectl port-forward service/postgres-rw :5432 </dev/null 2>&1
    port_forward_pid=$COPROC_PID
    while IFS='' read -r -t $PORT_FORWARD_TIMEOUT_SECONDS -u "${COPROC[0]}" LINE
    do
        if [[ "$LINE" == *"Forwarding from"* ]]; then
            port_forward_local_port="${LINE#Forwarding from 127.0.0.1:}"
            port_forward_local_port="${port_forward_local_port%% -> *}"
            if [ -z "${port_forward_local_port}" ]; then
              echo "ERROR: Failed to get local address for port-forward"
              echo "kubectl output line: $LINE"
              exit 1
            fi
            # Remaining output is on stderr, which we don't capture, so we
            # should be fine to ignore the stdout file descriptor now and
            # port_forward_pid remains set and will be used on cleanup
            #
            return
        else
            echo "kubectl: ${LINE}"
        fi
    done
    # if we reached here, read failed, likely due to the coproc exiting
    if [ -n "${port_forward_pid:-}" ]; then
      port_forward_ecode=
      wait $port_forward_pid || port_forward_ecode=$?
      echo "port-forward request failed? Exit code $port_forward_ecode"
    else
      echo "port forward request failed? Could not get kubectl port-forward's pid"
    fi
    exit 1
}

# Assumes there's only one coproc
function kill_port_forward() {
  if [ -n "${port_forward_pid:-}" ]; then
    kill ${port_forward_pid} || true
    wait -f ${port_forward_pid} || true
  fi
  port_foward_pid=
}

trap kill_port_forward EXIT

USER=$(kubectl get secret postgres-app -ojson | jq -r '.data.username' | base64 -d)
PASS=$(kubectl get secret postgres-app -ojson | jq -r '.data.password' | base64 -d)

start_port_forward

PGPASSWORD=$PASS psql -h localhost -p $port_forward_local_port -U $USER otf
