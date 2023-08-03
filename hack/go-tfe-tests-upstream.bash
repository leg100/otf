#!/usr/bin/env bash

# Run latest upstream go-tfe tests against tfe. The main bash script defaults
# to running tests against our fork of the go-tfe repo, but we are half-way
# through transitioning from our fork to the latest upstream.

set -e

function join_by { local IFS="$1"; shift; echo "$*"; }

export GO_TFE_REPO=github.com/hashicorp/go-tfe@latest
export TFE_ADDRESS="${TFE_ADDRESS:-https://localhost:8080}"

# go-tfe tests perform privileged operations (e.g. creating organizations), so
# we use a site admin token.
#
# NOTE: this token is the same token that otfd is configured with in docker
# compose
export TFE_TOKEN=${TFE_TOKEN:-site-token}
# skip testing paid tfe features
export SKIP_PAID=1
# use same self-signed certs used in docker-compose
export SSL_CERT_FILE=$PWD/internal/integration/fixtures/cert.pem

tests=()
tests+=('TestOrganizations')
tests+=('TestOrganizationTagsList/with_no_query_params')
tests+=('TestOrganizationTagsList/with_no_param_Filter')
# TODO: uncomment this once support is added for tag substring querying
#tests+=('TestOrganizationTagsList/with_no_param_Query')
tests+=('TestOrganizationTagsDelete')
tests+=('TestOrganizationTagsAddWorkspace')
tests+=('TestOrganizationTokens')
tests+=('TestWorkspacesCreate')
tests+=('TestWorkspacesUpdateByID')
tests+=('TestWorkspacesDelete')
tests+=('TestWorkspacesLock')
tests+=('TestWorkspacesUnlock')
tests+=('TestWorkspacesForceUnlock')
tests+=('TestWorkspaces_(Add|Remove)Tags')
tests+=('TestWorkspacesList/when_searching_using_a_tag')
tests+=('TestWorkspacesList/without_list_options')
tests+=('TestWorkspacesList/with_list_options')
tests+=('TestWorkspacesList/when_searching_a_known_workspace')
tests+=('TestWorkspacesList/when_searching_using_a_tag')
tests+=('TestWorkspacesList/when_searching_an_unknown_workspace')
tests+=('TestWorkspacesList/without_a_valid_organization')
tests+=('TestWorkspacesList/with_organization_included')
tests+=('TestWorkspacesRead/when_the_workspace_exists')
tests+=('TestWorkspacesRead/when_the_workspace_does_not_exist')
tests+=('TestWorkspacesRead/when_the_organization_does_not_exist')
tests+=('TestWorkspacesRead/without_a_valid_organization')
tests+=('TestWorkspacesRead/without_a_valid_workspace')
tests+=('TestWorkspacesReadByID')
tests+=('TestRunsList')
tests+=('TestNotificationConfigurationCreate/with_a')
tests+=('TestNotificationConfigurationCreate/without_a')
tests+=('TestNotificationConfigurationDelete')
tests+=('TestNotificationConfigurationUpdate/with_options')
tests+=('TestNotificationConfigurationUpdate/without_options')
tests+=('TestNotificationConfigurationUpdate/^when')
tests+=('TestTeamMembersAddByUsername')
tests+=('TestTeamMembersRemoveByUsernames')
tests+=('TestTeamMembersList')
tests+=('TestOAuthClientsCreate$')
tests+=('TestOAuthClientsRead')
tests+=('TestOAuthClientsList')
tests+=('TestOAuthClientsDelete')
tests+=('TestTeamsList')
tests+=('TestTeamsCreate')
tests+=('TestTeamsRead')
tests+=('TestTeamsUpdate$')
tests+=('TestTeamsDelete')
tests+=('TestConfigurationVersionsList')
tests+=('TestConfigurationVersionsCreate')
tests+=('TestConfigurationVersionsRead')
tests+=('TestConfigurationVersionsUpload')
tests+=('TestConfigurationVersionsDownload')
tests+=('TestVariables')
tests+=('TestStateVersion')
all=$(join_by '|' "${tests[@]}")

./hack/go-tfe-tests.bash $all
