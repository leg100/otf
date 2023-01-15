package otf

import "context"

type FakeWorkspaceLockService struct {
	WorkspaceLockService
}

func (f *FakeWorkspaceLockService) LockWorkspace(context.Context, string, WorkspaceLockOptions) (*Workspace, error) {
	return nil, nil
}
