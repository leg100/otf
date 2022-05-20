package otf

import "context"

var _ RunService = (*fakeRunService)(nil)

type fakeRunService struct {
	// mapping of run id to run
	db []*Run
	// list of ids of started runs
	started []string
	// prevent compiler error
	RunService
}

func newFakeRunService(runs ...*Run) *fakeRunService {
	return &fakeRunService{db: runs}
}

func (s *fakeRunService) List(_ context.Context, opts RunListOptions) (*RunList, error) {
	var items []*Run
	for _, r := range s.db {
		if *opts.WorkspaceID != r.Workspace.ID {
			continue
		}
		// if statuses are specified then run must match one of them.
		if len(opts.Statuses) > 0 && !ContainsRunStatus(opts.Statuses, r.Status) {
			continue
		}
		items = append(items, r)
	}
	return &RunList{
		Items: items,
	}, nil
}

func (s *fakeRunService) Start(_ context.Context, id string) error {
	s.started = append(s.started, id)
	return nil
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
