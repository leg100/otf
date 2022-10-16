package inmem

import (
	"context"

	"github.com/leg100/otf"
)

type fakeMapperApp struct {
	workspaces []*otf.Workspace
	runs       []*otf.Run
	events     chan otf.Event
	otf.Application
}

func (f *fakeMapperApp) ListWorkspaces(ctx context.Context, opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	return &otf.WorkspaceList{
		Items:      f.workspaces,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.workspaces)),
	}, nil
}

func (f *fakeMapperApp) ListRuns(ctx context.Context, opts otf.RunListOptions) (*otf.RunList, error) {
	return &otf.RunList{
		Items:      f.runs,
		Pagination: otf.NewPagination(opts.ListOptions, len(f.runs)),
	}, nil
}

func (f *fakeMapperApp) Watch(ctx context.Context, opts otf.WatchOptions) (<-chan otf.Event, error) {
	return f.events, nil
}
