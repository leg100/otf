package main

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTeam_NewCommand(t *testing.T) {
	team := &auth.Team{Name: "owners", Organization: "acme-corp"}
	cmd := fakeApp(withTeam(team)).teamNewCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created team owners\n", got.String())
}

func TestTeam_DeleteCommand(t *testing.T) {
	team := &auth.Team{Name: "owners", Organization: "acme-corp"}
	cmd := fakeApp(withTeam(team)).teamDeleteCommand()

	cmd.SetArgs([]string{"owners", "--organization", "acme-corp"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted team owners\n", got.String())
}
