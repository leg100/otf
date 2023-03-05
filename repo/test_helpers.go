package repo

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf"
	"github.com/leg100/otf/cloud"
	"github.com/stretchr/testify/require"
)

func newTestHook(t *testing.T, f factory, cloudID *string) *hook {
	want, err := f.newHook(newHookOpts{
		id:         otf.UUID(uuid.New()),
		secret:     otf.String("top-secret"),
		identifier: "leg100/" + uuid.NewString(),
		cloud:      "github",
		cloudID:    cloudID,
	})
	require.NoError(t, err)
	return want
}

func newTestFactory(t *testing.T, event cloud.VCSEvent) factory {
	return newFactory(
		fakeHostnameService{},
		fakeCloudService{event: event},
	)
}

type fakeCloudService struct {
	event cloud.VCSEvent
	cloud.Service
}

func (f fakeCloudService) GetCloudConfig(string) (cloud.Config, error) {
	return cloud.Config{Cloud: &fakeCloud{event: f.event}}, nil
}

type fakeCloud struct {
	event cloud.VCSEvent

	cloud.Cloud
}

func (f *fakeCloud) HandleEvent(w http.ResponseWriter, r *http.Request, opts cloud.HandleEventOptions) cloud.VCSEvent {
	return f.event
}

type fakeHostnameService struct {
	hostname string
	otf.HostnameService
}

func (f fakeHostnameService) Hostname() string { return f.hostname }

type fakeCloudClient struct {
	hook      cloud.Webhook // seed cloud with hook
	gotUpdate bool

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

func (f *fakeCloudClient) UpdateWebhook(context.Context, cloud.UpdateWebhookOptions) error {
	f.gotUpdate = true

	return nil
}

type fakeDB struct {
	hook *hook
	db
}

func (db *fakeDB) get(context.Context, uuid.UUID) (*hook, error) {
	return db.hook, nil
}
