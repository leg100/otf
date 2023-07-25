#!/usr/bin/env bash

# Run latest upstream go-tfe tests against tfe. The main bash script defaults
# to running tests against our fork of the go-tfe repo, but we are half-way
# through transitioning from our fork to the latest upstream.

set -e

function join_by { local IFS="$1"; shift; echo "$*"; }

export GO_TFE_REPO=github.com/hashicorp/go-tfe@latest

# necessary for TestOAuthClients* tests
export OAUTH_CLIENT_GITHUB_TOKEN="my-secret-github-token"

tests=()
tests+=('TestOrganizations')
tests+=('TestOrganizationTagsList/with_no_query_params')
tests+=('TestOrganizationTagsList/with_no_param_Filter')
# TODO: uncomment this once support is added for tag substring querying
#tests+=('TestOrganizationTagsList/with_no_param_Query')
tests+=('TestOrganizationTagsDelete')
tests+=('TestOrganizationTagsAddWorkspace')
tests+=('TestOrganizationTokens')
tests+=('TestWorkspaces_(Add|Remove)Tags')
tests+=('TestWorkspacesList/when_searching_using_a_tag')
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
