#!/usr/bin/env bash
#
# Creates a configuration version and uploads a tarball to it.
# Requires an organization and workspace name, and path to workspace
#

set -ex

ORG_NAME=$1
WORKSPACE_NAME=$2
WORKSPACE_PATH=$3

if [[ $# -ne 3 ]]
then
    echo "must provide org and workspace name"
    exit 1
fi

# oTF API doesn't yet verify tokens
TOKEN=dummy

HOST=localhost:8080

# Look up workspace ID
WORKSPACE_ID=$(curl -sS \
    -H "Authorization: Bearer $TOKEN" -H 'Accept: application/vnd.api+json' \
    https://$HOST/api/v2/organizations/$ORG_NAME/workspaces/$WORKSPACE_NAME \
    | jq -r '.data.id')

# Create configuration version
UPLOAD_URL=$(curl -sS \
    -H "Authorization: Bearer $TOKEN" -H 'Accept: application/vnd.api+json' \
    --request POST \
    --data '{"data":{"type":"configuration-versions"}}' \
    https://$HOST/api/v2/workspaces/$WORKSPACE_ID/configuration-versions \
    | jq -r '.data.attributes."upload-url"')

# Make tarball
TARBALL_PATH=$(mktemp)
tar zcf $TARBALL_PATH ./main.tf

# Upload configuration version
curl -sS \
    -H "Content-Type: application/octet-stream" \
    --request PUT \
    --data-binary @$TARBALL_PATH \
    https://$HOST$UPLOAD_URL
