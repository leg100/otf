package cli

import (
	"bytes"
	"testing"

	"github.com/leg100/otf/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserNewCommand(t *testing.T) {
	user := &auth.User{Username: "bobby"}
	cmd := fakeApp(withUser(user)).userNewCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created user bobby\n", got.String())
}

func TestUserDeleteCommand(t *testing.T) {
	cmd := fakeApp().userDeleteCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted user bobby\n", got.String())
}
