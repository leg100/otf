#!/usr/bin/env bash

# Setup tunnel to otfd service on local kind kubernetes cluster and open
# browser to connect to web app on local endpoint.

set -e

kubectl -n adjusted-ringtail port-forward services/otfd :80 
