package auth

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_teamNewCommand(t *testing.T) {
	team := &Team{Name: "owners", Organization: "acme-corp"}
	cmd := newFakeTeamCLI(team).teamNewCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created team owners\n", got.String())
}

func TestTeam_DeleteCommand(t *testing.T) {
	team := &Team{Name: "owners", Organization: "acme-corp"}
	cmd := newFakeTeamCLI(team).teamDeleteCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted team owners\n", got.String())
}

func TestTeam_AddMembership(t *testing.T) {
	team := &Team{Name: "owners", Organization: "acme-corp"}
	cmd := newFakeTeamCLI(team).addTeamMembershipCommand()

	cmd.SetArgs([]string{"bobby", "sally", "--organization", "acme-corp", "--team", "owners"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully added [bobby sally] to owners\n", got.String())
}

func TestTeam_RemoveMembership(t *testing.T) {
	team := &Team{Name: "owners", Organization: "acme-corp"}
	cmd := newFakeTeamCLI(team).deleteTeamMembershipCommand()

	cmd.SetArgs([]string{"bobby", "sally", "--organization", "acme-corp", "--team", "owners"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully removed [bobby sally] from owners\n", got.String())
}

type fakeTeamCLIService struct {
	team *Team
	AuthService
}

func newFakeTeamCLI(team *Team) *TeamCLI {
	return &TeamCLI{AuthService: &fakeTeamCLIService{team: team}}
}

func (f *fakeTeamCLIService) CreateTeam(context.Context, string, CreateTeamOptions) (*Team, error) {
	return f.team, nil

}

func (f *fakeTeamCLIService) GetTeam(context.Context, string, string) (*Team, error) {
	return f.team, nil
}

func (f *fakeTeamCLIService) DeleteTeam(context.Context, string) error {
	return nil
}

func (f *fakeTeamCLIService) AddTeamMembership(context.Context, TeamMembershipOptions) error {
	return nil
}

func (f *fakeTeamCLIService) RemoveTeamMembership(context.Context, TeamMembershipOptions) error {
	return nil
}
