package user

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/internal/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserNewCommand(t *testing.T) {
	cli := &userCLI{
		UserService: &fakeService{
			user: &User{Username: "bobby"},
		},
	}
	cmd := cli.userNewCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created user bobby\n", got.String())
}

func TestUserDeleteCommand(t *testing.T) {
	cli := &userCLI{
		UserService: &fakeService{},
	}
	cmd := cli.userDeleteCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted user bobby\n", got.String())
}

func TestTeam_AddMembership(t *testing.T) {
	cli := &membershipCLI{
		UserService: &fakeService{},
		TeamService: &fakeTeamService{
			team: &team.Team{Name: "owners", Organization: "acme-corp"},
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
		UserService: &fakeService{},
		TeamService: &fakeTeamService{
			team: &team.Team{Name: "owners", Organization: "acme-corp"},
		},
	}
	cmd := cli.deleteTeamMembershipCommand()

	cmd.SetArgs([]string{"bobby", "sally", "--organization", "acme-corp", "--team", "owners"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully removed [bobby sally] from owners\n", got.String())
}
