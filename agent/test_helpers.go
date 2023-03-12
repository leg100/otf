package agent

import (
	"context"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/otf/auth"
	"github.com/leg100/otf/client"
	"github.com/leg100/otf/run"
	"github.com/leg100/otf/variable"
	"github.com/leg100/otf/workspace"
	"github.com/stretchr/testify/require"
)

type fakeEnvironmentServices struct {
	ws *workspace.Workspace

	client.Client
}

func (f *fakeEnvironmentServices) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return f.ws, nil
}

func (f *fakeEnvironmentServices) CreateRegistrySession(context.Context, string) (*auth.RegistrySession, error) {
	return &auth.RegistrySession{}, nil
}

func (f *fakeEnvironmentServices) ListVariables(context.Context, string) ([]*variable.Variable, error) {
	return nil, nil
}

func (f *fakeEnvironmentServices) Hostname() string { return "fake-host.org" }

func newTestEnvironment(t *testing.T, ws *workspace.Workspace) *Environment {
	env, err := NewEnvironment(
		context.Background(),
		logr.Discard(),
		&fakeEnvironmentServices{ws: ws},
		&run.Run{},
		nil,
		nil,
		Config{},
	)
	require.NoError(t, err)
	return env
}
