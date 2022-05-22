package otf

import (
	"context"
	"fmt"
)

var _ RunService = (*fakeRunService)(nil)

type fakeRunService struct {
	// database of existing runs
	db []*Run
	// list of started runs
	started []*Run
	// prevent compiler error
	RunService
}

func newFakeRunService(runs ...*Run) *fakeRunService {
	return &fakeRunService{db: runs}
}

func (s *fakeRunService) List(_ context.Context, opts RunListOptions) (*RunList, error) {
	var items []*Run
	for _, r := range s.db {
		if opts.WorkspaceID != nil && *opts.WorkspaceID != r.Workspace.ID {
			continue
		}
		// if statuses are specified then run must match one of them.
		if len(opts.Statuses) > 0 && !ContainsRunStatus(opts.Statuses, r.Status()) {
			continue
		}
		items = append(items, r)
	}
	return &RunList{
		Items: items,
	}, nil
}

func (s *fakeRunService) Start(_ context.Context, id string) (*Run, error) {
	for _, r := range s.db {
		if r.ID == id {
			s.started = append(s.started, r)
			r.status = RunPlanQueued
			return r, nil
		}
	}
	return nil, fmt.Errorf("no run to start!")
}

type fakeWorkspaceService struct {
	workspaces []*Workspace
	WorkspaceService
}

func (s *fakeWorkspaceService) List(_ context.Context, opts WorkspaceListOptions) (*WorkspaceList, error) {
	return &WorkspaceList{
		Items: s.workspaces,
	}, nil
}
