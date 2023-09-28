package repo

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/cloud"
	"github.com/pkg/errors"
)

type (
	// synchroniser synchronises a hook with the vcs provider
	synchroniser struct {
		logr.Logger
		syncdb
	}

	syncdb interface {
		updateHookCloudID(ctx context.Context, id uuid.UUID, cloudID string) error
	}
)

// sync should be called from within a tx to avoid inconsistent results.
func (s *synchroniser) sync(ctx context.Context, client cloud.Client, hook *Hook) error {
	createAndSync := func() error {
		cloudID, err := client.CreateWebhook(ctx, cloud.CreateWebhookOptions{
			Repo:     hook.identifier,
			Secret:   hook.secret,
			Events:   defaultEvents,
			Endpoint: hook.endpoint,
		})
		if err != nil {
			return err
		}
		s.Info("created webhook", "webhook", hook)
		if err := s.updateHookCloudID(ctx, hook.id, cloudID); err != nil {
			return err
		}
		return nil
	}
	if hook.cloudID == nil {
		return createAndSync()
	}
	cloudHook, err := client.GetWebhook(ctx, cloud.GetWebhookOptions{
		Repo: hook.identifier,
		ID:   *hook.cloudID,
	})
	if errors.Is(err, internal.ErrResourceNotFound) {
		return createAndSync()
	} else if err != nil {
		return fmt.Errorf("retrieving hook from cloud: %w", err)
	}
	// hook is present on the vcs repo, but we update it anyway just to ensure
	// its configuration is consistent with what we have in the DB
	err = client.UpdateWebhook(ctx, cloudHook.ID, cloud.UpdateWebhookOptions{
		Repo:     hook.identifier,
		Secret:   hook.secret,
		Events:   defaultEvents,
		Endpoint: hook.endpoint,
	})
	if err != nil {
		return err
	}
	s.Info("updated webhook", "webhook", hook)
	return nil
}
