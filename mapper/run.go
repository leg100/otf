package mapper

import (
	"context"

	"github.com/leg100/otf"
)

var _ otf.RunService = (*RunMapper)(nil)

type workspaceName struct {
	name         string
	organization string
}

type RunMapper struct {
	// map from workspace name to workspace id
	workspaceNameMapper map[workspaceName]string

	// map run id to workspace id
	runMapper map[string]string

	// TODO: assign downstream run service (what to call it?)
	otf.RunService
}

func (s RunMapper) Create(ctx context.Context, spec otf.WorkspaceSpec, opts otf.RunCreateOptions) (*otf.Run, error) {
	run, err := s.RunService.Create(ctx, spec, opts)
	if err != nil {
		return nil, err
	}
	s.runMapper[run.ID()] = run.WorkspaceID()
	return run, nil
}

// Delete deletes a terraform run.
func (s RunMapper) Delete(ctx context.Context, runID string) error {
	err := s.RunService.Delete(ctx, runID)
	if err != nil {
		return err
	}
	delete(s.runMapper, runID)
	return nil
}
