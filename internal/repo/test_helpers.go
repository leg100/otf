package repo

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/stretchr/testify/require"
)

type (
	fakeCloudClient struct {
		hook      cloud.Webhook // seed cloud with hook
		gotUpdate bool

		cloud.Client
	}
	fakeDB struct {
		hook *Hook
	}
)

func newTestHook(t *testing.T, f factory, vcsProviderID string, cloudID *string) *Hook {
	want, err := f.newHook(newHookOptions{
		id:            internal.UUID(uuid.New()),
		vcsProviderID: vcsProviderID,
		secret:        internal.String("top-secret"),
		identifier:    "leg100/" + uuid.NewString(),
		cloud:         cloud.GithubKind,
		cloudID:       cloudID,
	})
	require.NoError(t, err)
	return want
}

func newTestFactory(t *testing.T, event cloud.VCSEvent) factory {
	return factory{
		HostnameService: internal.NewHostnameService(""),
	}
}

func (f *fakeCloudClient) CreateWebhook(context.Context, cloud.CreateWebhookOptions) (string, error) {
	return f.hook.ID, nil
}

func (f *fakeCloudClient) GetWebhook(ctx context.Context, opts cloud.GetWebhookOptions) (cloud.Webhook, error) {
	if f.hook.ID == opts.ID {
		return f.hook, nil
	}
	return cloud.Webhook{}, internal.ErrResourceNotFound
}

func (f *fakeCloudClient) UpdateWebhook(context.Context, string, cloud.UpdateWebhookOptions) error {
	f.gotUpdate = true

	return nil
}

func (f *fakeDB) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	f.hook.cloudID = &cloudID
	return nil
}
