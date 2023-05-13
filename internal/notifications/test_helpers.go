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
		workspace.WorkspaceService
	}
	// fakeFactory makes fake clients
	fakeFactory struct {
		published chan *run.Run
	}
	fakeClient struct {
		published chan *run.Run
	}
)

func newTestNotifier(t *testing.T, f clientFactory, configs ...*Config) *notifier {
	return &notifier{
		WorkspaceService: &fakeWorkspaceService{},
		cache:            newTestCache(t, f, configs...),
	}
}

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

func newTestConfig(t *testing.T, dst Destination, url string) *Config {
	cfg, err := NewConfig(uuid.NewString(), CreateConfigOptions{
		Name:            internal.String(uuid.NewString()),
		DestinationType: dst,
		Enabled:         internal.Bool(true),
		URL:             internal.String(url),
	})
	require.NoError(t, err)
	return cfg
}

func (db *fakeCacheDB) listAll(context.Context) ([]*Config, error) {
	return db.configs, nil
}

func (db *fakeWorkspaceService) GetWorkspace(context.Context, string) (*workspace.Workspace, error) {
	return nil, nil
}

func (f *fakeFactory) newClient(cfg *Config) (client, error) {
	return &fakeClient{f.published}, nil
}

func (f *fakeClient) Publish(r *run.Run, ws *workspace.Workspace) error {
	f.published <- r
	return nil
}

func (f *fakeClient) Close() {}
