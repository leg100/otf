package agent

import (
	"context"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/tokens"
)

const (
	JobTokenKind          tokens.Kind = "job_token"
	defaultJobTokenExpiry             = 60 * time.Minute
)

// createJobToken constructs a job token
func (f *tokenFactory) createJobToken(_ context.Context, spec JobSpec) ([]byte, error) {
	expiry := internal.CurrentTimestamp(nil).Add(defaultJobTokenExpiry)
	return f.NewToken(tokens.NewTokenOptions{
		Subject: spec.String(),
		Kind:    JobTokenKind,
		Expiry:  &expiry,
	})
}
