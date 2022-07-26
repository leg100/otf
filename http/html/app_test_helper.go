package html

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.Application = (*fakeApp)(nil)

type fakeApp struct {
	*fakeOrganizationService
	*fakeWorkspaceService
	*fakeRunService
	*fakeUserService

	// TODO: stubbed until tests are implemented
	otf.StateVersionService
	otf.ConfigurationVersionService
	otf.EventService
}

type fakeUserService struct {
	fakeUser *otf.User
	otf.UserService
}

func (u *fakeUserService) GetUser(context.Context, otf.UserSpec) (*otf.User, error) {
	return u.fakeUser, nil
}

func (u *fakeUserService) CreateToken(ctx context.Context, user *otf.User, opts *otf.TokenCreateOptions) (*otf.Token, error) {
	return otf.NewToken(user.ID(), opts.Description)
}

func (u *fakeUserService) DeleteToken(context.Context, *otf.User, string) error { return nil }

type fakeOrganizationService struct {
	fakeOrganization *otf.Organization
	otf.OrganizationService
}

func (u *fakeOrganizationService) GetOrganization(ctx context.Context, name string) (*otf.Organization, error) {
	return u.fakeOrganization, nil
}

func (u *fakeOrganizationService) ListOrganizations(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return &otf.OrganizationList{
		Items:      []*otf.Organization{u.fakeOrganization},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}, nil
}

type fakeWorkspaceService struct {
	fakeWorkspace *otf.Workspace
	otf.WorkspaceService
}

func (u *fakeWorkspaceService) GetWorkspace(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

func (u *fakeWorkspaceService) UpdateWorkspace(context.Context, otf.WorkspaceSpec, otf.WorkspaceUpdateOptions) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

func (u *fakeWorkspaceService) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      []*otf.Workspace{u.fakeWorkspace},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}, nil
}

func (u *fakeWorkspaceService) CreateWorkspace(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

type fakeRunService struct {
	fakeRun *otf.Run
	otf.RunService
}

func (u *fakeRunService) GetRun(context.Context, string) (*otf.Run, error) {
	return u.fakeRun, nil
}

func (u *fakeRunService) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      []*otf.Run{u.fakeRun},
		Pagination: otf.NewPagination(opts.ListOptions, 1),
	}, nil
}

func (u *fakeRunService) GetChunk(context.Context, string, otf.PhaseType, otf.GetChunkOptions) (otf.Chunk, error) {
	return otf.Chunk{Data: []byte("fake-logs")}, nil
}
