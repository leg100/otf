package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/team"
	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMembershipClient struct {
	team *team.Team
}

func (f *fakeMembershipClient) AddTeamMembership(context.Context, resource.TfeID, []user.Username) error {
	return nil
}

func (f *fakeMembershipClient) RemoveTeamMembership(context.Context, resource.TfeID, []user.Username) error {
	return nil
}

func (f *fakeMembershipClient) GetTeam(context.Context, organization.Name, string) (*team.Team, error) {
	return f.team, nil
}

func TestTeam_AddMembership(t *testing.T) {
	cli := &membershipCLI{
		client: &fakeMembershipClient{
			team: &team.Team{Name: "owners", Organization: organization.NewTestName(t)},
		},
	}
	cmd := cli.addTeamMembershipCommand()

	cmd.SetArgs([]string{"bobby", "sally", "--organization", "acme-corp", "--team", "owners"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully added [bobby sally] to owners\n", got.String())
}

func TestTeam_RemoveMembership(t *testing.T) {
	cli := &membershipCLI{
		client: &fakeMembershipClient{
			team: &team.Team{Name: "owners", Organization: organization.NewTestName(t)},
		},
	}
	cmd := cli.deleteTeamMembershipCommand()

	cmd.SetArgs([]string{"bobby", "sally", "--organization", "acme-corp", "--team", "owners"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully removed [bobby sally] from owners\n", got.String())
}
