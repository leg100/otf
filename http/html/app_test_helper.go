package html

import (
	"context"

	"github.com/leg100/otf"
)

type fakeApp struct {
	otf.Application
	fakeUserService         *fakeUserService
	fakeOrganizationService *fakeOrganizationService
	fakeWorkspaceService    *fakeWorkspaceService
	fakeRunService          *fakeRunService
}

func (a fakeApp) UserService() otf.UserService {
	return a.fakeUserService
}

func (a fakeApp) OrganizationService() otf.OrganizationService {
	return a.fakeOrganizationService
}

func (a fakeApp) WorkspaceService() otf.WorkspaceService {
	return a.fakeWorkspaceService
}

func (a fakeApp) RunService() otf.RunService {
	return a.fakeRunService
}

type fakeUserService struct {
	fakeUser *otf.User
	otf.UserService
}

func (u *fakeUserService) Get(context.Context, otf.UserSpec) (*otf.User, error) {
	return u.fakeUser, nil
}

type fakeOrganizationService struct {
	fakeOrganization *otf.Organization
	otf.OrganizationService
}

func (u *fakeOrganizationService) Get(ctx context.Context, name string) (*otf.Organization, error) {
	return u.fakeOrganization, nil
}

func (u *fakeOrganizationService) List(ctx context.Context, opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	return &otf.OrganizationList{
		Items: []*otf.Organization{u.fakeOrganization},
	}, nil
}

type fakeWorkspaceService struct {
	fakeWorkspace *otf.Workspace
	otf.WorkspaceService
}

func (u *fakeWorkspaceService) Get(ctx context.Context, spec otf.WorkspaceSpec) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

func (u *fakeWorkspaceService) List(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items: []*otf.Workspace{u.fakeWorkspace},
	}, nil
}

func (u *fakeWorkspaceService) Create(ctx context.Context, opts otf.WorkspaceCreateOptions) (*otf.Workspace, error) {
	return u.fakeWorkspace, nil
}

type fakeRunService struct {
	fakeRun *otf.Run
	otf.RunService
}

func (u *fakeRunService) Get(context.Context, string) (*otf.Run, error) {
	return u.fakeRun, nil
}

func (u *fakeRunService) List(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items: []*otf.Run{u.fakeRun},
	}, nil
}
