package otf

import "context"

type FakeWorkspaceLockService struct {
	WorkspaceLockService
}

func (f *FakeWorkspaceLockService) LockWorkspace(context.Context, WorkspaceSpec, WorkspaceLockOptions) (*Workspace, error) {
	return nil, nil
}
