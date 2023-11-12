package team

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_teamNewCommand(t *testing.T) {
	cli := &teamCLI{
		TeamService: &fakeService{
			team: &Team{Name: "owners", Organization: "acme-corp"},
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
		TeamService: &fakeService{
			team: &Team{Name: "owners", Organization: "acme-corp"},
		},
	}
	cmd := cli.teamDeleteCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted team owners\n", got.String())
}
