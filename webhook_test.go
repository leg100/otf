package otf

import (
	"context"
	"testing"

	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookSynchroniser(t *testing.T) {
	// input is the unsynchronised hook that should be used in each test case
	input, err := NewUnsynchronisedWebhook(NewUnsynchronisedWebhookOptions{
		Identifier: "leg100/otf",
		Cloud:      "github",
	})
	require.NoError(t, err)

	host := "fake-host.org:12345"

	tests := []struct {
		name  string
		db    *Webhook      // seed DB with hook
		cloud cloud.Webhook // seed cloud with hook
		want  *Webhook      // synchronised hook
	}{
		{
			name: "no changes",
			db:   &Webhook{cloudID: "123", UnsynchronisedWebhook: input},
			cloud: cloud.Webhook{
				ID:         "123",
				Identifier: input.identifier,
				Events:     DefaultWebhookEvents,
				Endpoint:   input.Endpoint(host),
			},
			want: &Webhook{cloudID: "123", UnsynchronisedWebhook: input},
		},
		{
			name:  "new hook",
			cloud: cloud.Webhook{ID: "new-cloud-id"}, // new id that cloud returns
			want:  &Webhook{cloudID: "new-cloud-id", UnsynchronisedWebhook: input},
		},
		{
			name:  "hook in DB but not on cloud",
			db:    &Webhook{cloudID: "123", UnsynchronisedWebhook: input},
			cloud: cloud.Webhook{ID: "new-cloud-id"}, // new id that cloud returns
			want:  &Webhook{cloudID: "new-cloud-id", UnsynchronisedWebhook: input},
		},
		{
			name: "hook events missing on cloud",
			db:   &Webhook{cloudID: "123", UnsynchronisedWebhook: input},
			cloud: cloud.Webhook{
				ID:         "123",
				Identifier: input.identifier,
				Endpoint:   input.Endpoint(host),
			},
			want: &Webhook{cloudID: "123", UnsynchronisedWebhook: input},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synchr := &WebhookSynchroniser{
				HostnameService:    fakeWebhookSynchroniserHostnameService{},
				DB:                 &fakeWebhookSynchroniserDB{hook: tt.db},
				VCSProviderService: &fakeWebhookSynchroniserProviderService{hook: tt.cloud},
			}
			got, err := synchr.Synchronise(context.Background(), "", input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

type fakeWebhookSynchroniserDB struct {
	hook *Webhook

	DB
}

func (f *fakeWebhookSynchroniserDB) SynchroniseWebhook(ctx context.Context, unsynced *UnsynchronisedWebhook, cb func(*Webhook) (string, error)) (*Webhook, error) {
	// if db has been seeded with an existing hook, then return that along with
	// cloud ID; otherwise return the unsynced hook the caller passed in.
	cloudID, err := cb(f.hook)
	if f.hook == nil {
		return &Webhook{cloudID: cloudID, UnsynchronisedWebhook: unsynced}, err
	}
	f.hook.cloudID = cloudID
	return f.hook, nil
}

type fakeWebhookSynchroniserProviderService struct {
	hook cloud.Webhook

	VCSProviderService
}

func (f *fakeWebhookSynchroniserProviderService) GetVCSClient(context.Context, string) (cloud.Client, error) {
	return &fakeWebhookSynchroniserClient{hook: f.hook}, nil
}

type fakeWebhookSynchroniserClient struct {
	hook cloud.Webhook

	cloud.Client
}

func (f *fakeWebhookSynchroniserClient) CreateWebhook(context.Context, cloud.CreateWebhookOptions) (string, error) {
	return f.hook.ID, nil
}

func (f *fakeWebhookSynchroniserClient) GetWebhook(ctx context.Context, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	if f.hook.ID == opts.ID {
		return f.hook, nil
	}
	return cloud.Webhook{}, ErrResourceNotFound
}

func (*fakeWebhookSynchroniserClient) UpdateWebhook(context.Context, cloud.UpdateWebhookOptions) error {
	return nil
}

type fakeWebhookSynchroniserHostnameService struct {
	HostnameService
}

func (fakeWebhookSynchroniserHostnameService) Hostname() string { return "fake-host.org:12345" }
