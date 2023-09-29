package repo

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
)

type (
	fakeCloudClient struct {
		hook      cloud.Webhook // seed cloud with hook
		gotUpdate bool

		cloud.Client
	}
	fakeDB struct {
		hook *hook
	}
)

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
