package otf

import (
	"context"

	"github.com/google/uuid"
	"github.com/leg100/otf/cloud"
)

type (
	// HookService hooks up, and unhooks, resources to webhooks.
	HookService interface {
		Hook(ctx context.Context, opts HookOptions) error
		Unhook(ctx context.Context, opts UnhookOptions) error
	}

	HookOptions struct {
		Identifier string
		Cloud      string

		HookCallback
		cloud.Client
	}
	HookCallback func(ctx context.Context, tx Database, hookID uuid.UUID) error

	UnhookOptions struct {
		HookID uuid.UUID

		UnhookCallback
		cloud.Client
	}
	UnhookCallback func(ctx context.Context, tx Database) error
)
