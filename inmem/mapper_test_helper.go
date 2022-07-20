package inmem

import (
	"context"

	"github.com/leg100/otf"
)

type fakeWorkspaceService struct {
	workspaces []*otf.Workspace

	otf.WorkspaceService
}

func (f *fakeWorkspaceService) ListWorkspace(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

type fakeRunService struct {
	runs []*otf.Run

	otf.RunService
}

func (f *fakeRunService) ListRun(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
}

type fakeSubject struct {
	// name of organization the subject is a member of
	memberOrg string
}

func (f *fakeSubject) CanAccess(org *string) bool {
	return *org == f.memberOrg
}
