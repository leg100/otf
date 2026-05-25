package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTeamClient struct {
	team *team.Team
}

func (f *fakeTeamClient) CreateTeam(context.Context, organization.Name, team.CreateTeamOptions) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamClient) GetTeam(context.Context, organization.Name, string) (*team.Team, error) {
	return f.team, nil
}

func (f *fakeTeamClient) DeleteTeam(context.Context, resource.TfeID) error {
	return nil
}

func Test_TeamNewCommand(t *testing.T) {
	cli := &teamCLI{
		client: &fakeTeamClient{
			team: &team.Team{Name: "owners", Organization: organization.NewTestName(t)},
		},
	}
	cmd := cli.teamNewCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created team owners\n", got.String())
}

func TestTeam_DeleteCommand(t *testing.T) {
	cli := &teamCLI{
		client: &fakeTeamClient{
			team: &team.Team{Name: "owners", Organization: organization.NewTestName(t)},
		},
	}
	cmd := cli.teamDeleteCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted team owners\n", got.String())
}
