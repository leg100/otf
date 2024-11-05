package repohooks

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/vcs"
)

type (
	fakeCloudClient struct {
		hook      vcs.Webhook // seed cloud with hook
		gotUpdate bool

		vcs.Client
	}
	fakeDB struct {
		hook *hook
	}
)

func (f *fakeCloudClient) CreateWebhook(context.Context, vcs.CreateWebhookOptions) (string, error) {
	return f.hook.ID, nil
}

func (f *fakeCloudClient) GetWebhook(ctx context.Context, opts vcs.GetWebhookOptions) (vcs.Webhook, error) {
	if f.hook.ID == opts.ID {
		return f.hook, nil
	}
	return vcs.Webhook{}, internal.ErrResourceNotFound
}

func (f *fakeCloudClient) UpdateWebhook(context.Context, string, vcs.UpdateWebhookOptions) error {
	f.gotUpdate = true

	return nil
}

func (f *fakeDB) updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error {
	f.hook.cloudID = &cloudID
	return nil
}
