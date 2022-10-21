package main

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspaceList(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws1 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	ws2 := otf.NewTestWorkspace(t, org, otf.WorkspaceCreateOptions{})
	factory := &fakeWorkspaceListClientFactory{workspaces: []*otf.Workspace{ws1, ws2}}

	cmd := WorkspaceListCommand(factory)
	cmd.SetArgs([]string{"--organization", org.Name()})
	got := bytes.Buffer{}
	cmd.SetOut(&got)
	require.NoError(t, cmd.Execute())
	want := fmt.Sprintf("%s\n%s\n", ws1.Name(), ws2.Name())
	assert.Equal(t, want, got.String())
}

func TestWorkspaceListMissingOrganization(t *testing.T) {
	cmd := WorkspaceListCommand(&fakeClientFactory{})
	cmd.SetArgs([]string{"automatize"})
	err := cmd.Execute()
	assert.EqualError(t, err, "required flag(s) \"organization\" not set")
}

type fakeWorkspaceListClientFactory struct {
	workspaces []*otf.Workspace
}

func (f fakeWorkspaceListClientFactory) NewClient() (otfhttp.Client, error) {
	return &fakeWorkspaceListClient{
		workspaces: f.workspaces,
	}, nil
}

type fakeWorkspaceListClient struct {
	workspaces []*otf.Workspace
	otf.Application
}

func (f *fakeWorkspaceListClient) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}
