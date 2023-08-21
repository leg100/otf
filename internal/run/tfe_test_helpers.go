package run

import (
	"context"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/workspace"
)

type (
	fakeRunService struct {
		ch chan pubsub.Event

		RunService
	}
	fakePermissionsService struct {
		workspace.PermissionsService
	}
)

func (f *fakeRunService) Watch(context.Context, WatchOptions) (<-chan pubsub.Event, error) {
	return f.ch, nil
}

func (f *fakePermissionsService) GetPolicy(ctx context.Context, workspaceID string) (internal.WorkspacePolicy, error) {
	return internal.WorkspacePolicy{}, nil
}
