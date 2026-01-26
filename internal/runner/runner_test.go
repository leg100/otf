package runner

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunner(t *testing.T) {
	updates := make(chan RunnerStatus)
	wantID := resource.NewTfeID(resource.RunnerKind)

	r, err := New(
		logr.Discard(),
		&fakeRunnerClient{registeredID: wantID, updates: updates},
		func(jobToken string) OperationClient {
			return OperationClient{}
		},
		NewDefaultConfig(),
	)
	require.NoError(t, err)

	// Terminate runner at end of test
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startErr := make(chan error)
	go func() {
		startErr <- r.Start(ctx)
	}()

	// Test that the runner registers itself
	<-r.registered
	assert.Equal(t, wantID, r.ID)
	// Terminate runner
	cancel()
	// Test that runner sends final status update
	assert.Equal(t, RunnerExited, <-updates)
}

type fakeRunnerClient struct {
	RunnerClient

	registeredID resource.TfeID
	updates      chan RunnerStatus
}

func (f *fakeRunnerClient) Register(ctx context.Context, opts RegisterRunnerOptions) (*RunnerMeta, error) {
	return &RunnerMeta{ID: f.registeredID}, nil
}

func (f *fakeRunnerClient) awaitAllocatedJobs(ctx context.Context, _ resource.TfeID) ([]*Job, error) {
	// Block until context canceled
	<-ctx.Done()
	return nil, nil
}

func (f *fakeRunnerClient) updateStatus(ctx context.Context, agentID resource.TfeID, status RunnerStatus) error {
	f.updates <- status
	return nil
}
