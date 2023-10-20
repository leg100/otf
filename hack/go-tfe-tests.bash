#!/usr/bin/env bash

# Runs go-tfe integration tests against OTF. The defaults assume you've started
# up the docker-compose stack. If not, you'll need to override the environment
# variables below. You need an API user token with site admin privileges. You
# can pass individual test names as arguments to override the default behaviour
# of running all tests.
#
# Read the docs first: https://docs.otf.ninja/latest/testing/

set -e

function join_by { local IFS="$1"; shift; echo "$*"; }

export TFE_ADDRESS="${TFE_ADDRESS:-https://localhost:8080}"
export TFE_TOKEN=${TFE_TOKEN:-site-token}
export SKIP_PAID=1
export SSL_CERT_FILE=$PWD/internal/integration/fixtures/cert.pem

tests=()
tests+=('TestOrganizations')
tests+=('TestOrganizationTagsList/with_no_query_params')
tests+=('TestOrganizationTagsList/with_no_param_Filter')
tests+=('TestOrganizationTagsDelete')
tests+=('TestOrganizationTagsAddWorkspace')
tests+=('TestOrganizationTokens')
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
tests+=('TestWorkspacesRead$/when_the_workspace_exists')
tests+=('TestWorkspacesRead$/when_the_workspace_does_not_exist')
tests+=('TestWorkspacesRead$/when_the_organization_does_not_exist')
tests+=('TestWorkspacesRead$/without_a_valid_organization')
tests+=('TestWorkspacesRead$/without_a_valid_workspace')
tests+=('TestWorkspacesReadByID')
tests+=('TestRunsCreate')
tests+=('TestRunsList')
tests+=('TestRunsCancel')
tests+=('TestRunsForceCancel')
tests+=('TestRunsDiscard')
tests+=('TestPlans')
tests+=('TestAppliesRead')
tests+=('TestAppliesLogs')
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
tests+=('TestTeamToken')
tests+=('TestConfigurationVersionsList')
tests+=('TestConfigurationVersionsCreate')
tests+=('TestConfigurationVersionsUpload')
tests+=('TestConfigurationVersionsDownload')
tests+=('TestVariables')
tests+=('TestVariableSetsCreate')
tests+=('TestVariableSetsUpdate$')
tests+=('TestVariableSetsList$')
tests+=('TestVariableSetsListForWorkspace')
tests+=('TestVariableSetsRead')
tests+=('TestVariableSetsApplyToAndRemoveFromWorkspaces')
tests+=('TestVariableSetsDelete')
tests+=('TestVariableSetVariables')
tests+=('TestStateVersion')

# only run these tests if env vars are present - otherwise the tests fail early
vcsTests=('TestConfigurationVersionsRead' 'TestWorkspacesCreate')
if [ -n "$GITHUB_POLICY_SET_IDENTIFIER" ] && [ -n "$OAUTH_CLIENT_GITHUB_TOKEN" ]
then
    tests=( "${tests[@]}" "${vcsTests[@]}" )
else
    echo "skipping ${vcsTests[@]} because GITHUB_POLICY_SET_IDENTIFIER and OAUTH_CLIENT_GITHUB_TOKEN are missing"
fi

all=$(join_by '|' "${tests[@]}")

dest_dir=$(go mod download -json github.com/hashicorp/go-tfe@latest | jq -r '.Dir')
echo "downloaded go-tfe module to $dest_dir"

# some tests generate a tarball and save locally and need write perms
chmod -R +w $dest_dir

cd $dest_dir
go test -v -run ${@:-$all} -timeout 600s
