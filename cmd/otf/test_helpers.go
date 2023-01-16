package main

import (
	"context"

	"github.com/leg100/otf"
	otfhttp "github.com/leg100/otf/http"
)

type fakeClientFactory struct {
	ws *otf.Workspace
}

func (f fakeClientFactory) NewClient() (otfhttp.Client, error) {
	return &fakeClient{
		ws: f.ws,
	}, nil
}

type fakeClient struct {
	ws *otf.Workspace
	otf.Application
}

func (f *fakeClient) CreateOrganization(ctx context.Context, opts otf.OrganizationCreateOptions) (*otf.Organization, error) {
	return otf.NewOrganization(opts)
}

func (f *fakeClient) GetWorkspace(context.Context, string) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeClient) GetWorkspaceByName(context.Context, string, string) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeClient) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      []*otf.Workspace{f.ws},
		Pagination: otf.NewPagination(otf.ListOptions{}, 1),
	}, nil
}

func (f *fakeClient) UpdateWorkspace(ctx context.Context, workspaceID string, opts otf.UpdateWorkspaceOptions) (*otf.Workspace, error) {
	f.ws.Update(opts)
	return f.ws, nil
}

func (f *fakeClient) LockWorkspace(context.Context, string, otf.WorkspaceLockOptions) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeClient) UnlockWorkspace(context.Context, string, otf.WorkspaceUnlockOptions) (*otf.Workspace, error) {
	return f.ws, nil
}

func (f *fakeClient) CreateAgentToken(ctx context.Context, opts otf.CreateAgentTokenOptions) (*otf.AgentToken, error) {
	return otf.NewAgentToken(opts)
}
