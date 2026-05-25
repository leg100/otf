package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/leg100/otf/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeClient struct {
	user *user.User
}

func (f *fakeClient) Create(context.Context, string, ...user.NewUserOption) (*user.User, error) {
	return f.user, nil
}

func (f *fakeClient) Delete(context.Context, user.Username) error {
	return nil
}

func TestUserNewCommand(t *testing.T) {
	cli := &userCLI{
		client: &fakeClient{
			user: &user.User{Username: user.MustUsername("bobby")},
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
		client: &fakeClient{},
	}
	cmd := cli.userDeleteCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted user bobby\n", got.String())
}
