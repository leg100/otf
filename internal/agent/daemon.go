package agent

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/logr"
)

const defaultConcurrency = 5

// daemon implements the agent itself.
type daemon struct {
	logr.Logger
}

// New constructs a new agent daemon.
func New(ctx context.Context, logger logr.Logger, app client, cfg Config) (*daemon, error) {
	if cfg.server {
		// Confirm token validity
		at, err := app.GetAgentToken(ctx, "")
		if err != nil {
			return nil, fmt.Errorf("attempted authentication: %w", err)
		}
		logger.Info("successfully authenticated", "organization", at.Organization, "token_id", at.ID)
	}
	return nil, nil
}
