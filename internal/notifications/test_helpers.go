package notifications

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/pubsub"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type (
	fakeCacheDB struct {
		configs []*Config
	}
	fakeWorkspaceService struct{}
	fakeRunService       struct{}
	fakeHostnameService  struct {
		*internal.HostnameService
	}
	// fakeFactory makes fake clients
	fakeFactory struct {
		published chan *notification
	}
	fakeClient struct {
		published chan *notification
	}
)

func newTestCache(t *testing.T, f clientFactory, configs ...*Config) *cache {
	if f == nil {
		f = &fakeFactory{}
	}
	cache, err := newCache(context.Background(),
		&fakeCacheDB{configs: configs},
		f,
	)
	require.NoError(t, err)
	return cache
}

func newTestConfig(t *testing.T, workspaceID resource.TfeID, dst Destination, url string, triggers ...Trigger) *Config {
	cfg, err := NewConfig(workspaceID, CreateConfigOptions{
		Name:            new(uuid.NewString()),
		DestinationType: dst,
		Enabled:         new(true),
		URL:             new(url),
		Triggers:        triggers,
	})
	require.NoError(t, err)
	return cfg
}

func (f *fakeCacheDB) listAll(context.Context) ([]*Config, error) {
	return f.configs, nil
}

func (f *fakeWorkspaceService) Get(context.Context, resource.TfeID) (*workspace.Workspace, error) {
	return nil, nil
}

func (f *fakeRunService) Get(ctx context.Context, id resource.TfeID) (*run.Run, error) {
	return &run.Run{ID: id}, nil
}

func (f *fakeRunService) Watch(ctx context.Context) (<-chan pubsub.Event[*run.Event], func()) {
	return nil, nil
}

func (*fakeHostnameService) Hostname() string { return "" }

func (f *fakeFactory) newClient(cfg *Config) (client, error) {
	return &fakeClient{f.published}, nil
}

func (f *fakeClient) Publish(ctx context.Context, n *notification) error {
	f.published <- n
	return nil
}

func (f *fakeClient) Close() {}
