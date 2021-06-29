package main

import (
	"context"

	"github.com/leg100/go-tfe"
)

type FakeClientConfig struct{}

func (f FakeClientConfig) NewClient() (Client, error) { return &FakeClient{}, nil }

type FakeClient struct {
	Client
}

func (f FakeClient) Organizations() tfe.Organizations { return &FakeOrganizationsClient{} }

func (f FakeClient) Workspaces() tfe.Workspaces { return &FakeWorkspacesClient{} }

type FakeOrganizationsClient struct {
	tfe.Organizations
}

func (f *FakeOrganizationsClient) Create(ctx context.Context, opts tfe.OrganizationCreateOptions) (*tfe.Organization, error) {
	return &tfe.Organization{
		Name:  *opts.Name,
		Email: *opts.Email,
	}, nil
}

type FakeWorkspacesClient struct {
	tfe.Workspaces
}

func (f *FakeWorkspacesClient) Read(ctx context.Context, org string, ws string) (*tfe.Workspace, error) {
	return &tfe.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Lock(ctx context.Context, id string, opts tfe.WorkspaceLockOptions) (*tfe.Workspace, error) {
	return &tfe.Workspace{
		ID: "ws-123",
	}, nil
}

func (f *FakeWorkspacesClient) Unlock(ctx context.Context, id string) (*tfe.Workspace, error) {
	return &tfe.Workspace{
		ID: "ws-123",
	}, nil
}
