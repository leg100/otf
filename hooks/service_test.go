package hooks

import (
	"context"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

func TestService_Synchronisation(t *testing.T) {
	// input is the unsynchronised hook that should be used in each test case
	input, err := newHook(newHookOpts{
		identifier: "leg100/otf",
		cloud:      "github",
	})
	require.NoError(t, err)

	tests := []struct {
		name  string
		db    *synced       // seed DB with hook
		cloud cloud.Webhook // seed cloud with hook
		want  *synced       // wanted end result
	}{
		{
			name: "no changes",
			db:   &synced{cloudID: "123", unsynced: input},
			cloud: cloud.Webhook{
				ID:         "123",
				Identifier: input.identifier,
				Events:     defaultEvents,
				Endpoint:   input.endpoint,
			},
			want: &synced{cloudID: "123", unsynced: input},
		},
		{
			name:  "new hook",
			cloud: cloud.Webhook{ID: "new-cloud-id"}, // new id that cloud returns
			want:  &synced{cloudID: "new-cloud-id", unsynced: input},
		},
		{
			name:  "hook in DB but not on cloud",
			db:    &synced{cloudID: "123", unsynced: input},
			cloud: cloud.Webhook{ID: "new-cloud-id"}, // new id that cloud returns
			want:  &synced{cloudID: "new-cloud-id", unsynced: input},
		},
		{
			name: "hook events missing on cloud",
			db:   &synced{cloudID: "123", unsynced: input},
			cloud: cloud.Webhook{
				ID:         "123",
				Identifier: input.identifier,
				Endpoint:   input.endpoint,
			},
			want: &synced{cloudID: "123", unsynced: input},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &Service{
				db:              &fakeDB{hook: tt.db},
				HostnameService: fakeHostnameService{},
			}
			err := svc.RegisterConnection(context.Background(), RegisterConnectionOptions{
				Identifier: input.identifier,
				Cloud:      input.cloud,
				Client:     &fakeCloudClient{hook: tt.cloud},
				Connection: fakeConnection{},
			})
			require.NoError(t, err)
		})
	}
}

type fakeConnection struct {
	Connection
}

func (f fakeConnection) Connect(otf.Database, string) error {
	return nil
}

type fakeDB struct {
	hook *synced

	db
}

func (f *fakeDB) create(ctx context.Context, unsynced *unsynced, fn syncFunc) (*synced, error) {
	cloudID, err := fn(f.hook, nil)

	// if db has been seeded with an existing hook, then return that along with
	// cloud ID; otherwise return the unsynced hook the caller passed in.
	if f.hook != nil {
		f.hook.cloudID = cloudID
		return f.hook, nil
	}
	return &synced{cloudID: cloudID, unsynced: unsynced}, err
}

type fakeCloudClient struct {
	hook cloud.Webhook

	cloud.Client
}

func (f *fakeCloudClient) CreateWebhook(context.Context, cloud.CreateWebhookOptions) (string, error) {
	return f.hook.ID, nil
}

func (f *fakeCloudClient) GetWebhook(ctx context.Context, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	if f.hook.ID == opts.ID {
		return f.hook, nil
	}
	return cloud.Webhook{}, otf.ErrResourceNotFound
}

func (*fakeCloudClient) UpdateWebhook(context.Context, cloud.UpdateWebhookOptions) error {
	return nil
}

type fakeHostnameService struct {
	otf.HostnameService
}

func (fakeHostnameService) Hostname() string { return "fake-host.org:12345" }
