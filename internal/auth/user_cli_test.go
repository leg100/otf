package auth

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserNewCommand(t *testing.T) {
	user := &User{Username: "bobby"}
	cmd := newFakeUserCLI(user).userNewCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully created user bobby\n", got.String())
}

func TestUserDeleteCommand(t *testing.T) {
	cmd := newFakeUserCLI(nil).userDeleteCommand()

	cmd.SetArgs([]string{"bobby"})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())

	assert.Equal(t, "Successfully deleted user bobby\n", got.String())
}

type fakeUserCLIService struct {
	user *User
	UserService
}

func newFakeUserCLI(user *User) *UserCLI {
	return &UserCLI{UserService: &fakeUserCLIService{user: user}}
}

func (f *fakeUserCLIService) CreateUser(context.Context, string, ...NewUserOption) (*User, error) {
	return f.user, nil
}

func (f *fakeUserCLIService) DeleteUser(context.Context, string) error {
	return nil
}
