package notifications

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/run"
	"github.com/leg100/otf/internal/workspace"
	"github.com/stretchr/testify/require"
)

type (
	fakeCacheDB struct {
		configs []*Config
	}
	fakeWorkspaceService struct {
		workspace.Service
	}
	fakeHostnameService struct {
		*internal.HostnameService
	}
	// fakeFactory makes fake clients
	fakeFactory struct {
		published chan *run.Run
	}
	fakeClient struct {
		published chan *run.Run
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

func newTestConfig(t *testing.T, workspaceID string, dst Destination, url string, triggers ...Trigger) *Config {
	cfg, err := NewConfig(workspaceID, CreateConfigOptions{
		Name:            internal.String(uuid.NewString()),
		DestinationType: dst,
		Enabled:         internal.Bool(true),
		URL:             internal.String(url),
		Triggers:        triggers,
	})
	require.NoError(t, err)
	return cfg
}

func (db *fakeCacheDB) listAll(context.Context) ([]*Config, error) {
	return db.configs, nil
}

func (db *fakeWorkspaceService) Get(context.Context, string) (*workspace.Workspace, error) {
	return nil, nil
}

func (db *fakeHostnameService) Hostname() string { return "" }

func (f *fakeFactory) newClient(cfg *Config) (client, error) {
	return &fakeClient{f.published}, nil
}

func (f *fakeClient) Publish(ctx context.Context, n *notification) error {
	f.published <- n.run
	return nil
}

func (f *fakeClient) Close() {}
