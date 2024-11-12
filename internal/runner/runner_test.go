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
	wantID := resource.NewID(resource.RunnerKind)

	r, err := newRunner(
		logr.Discard(),
		&fakeRunnerClient{registeredID: wantID, updates: updates},
		&fakeOperationSpawner{},
		false,
		Config{},
	)
	require.NoError(t, err)

	// Terminate runner at end of test
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	startErr := make(chan error)
	go func() {
		startErr <- r.Start(ctx)
	}()

	// Test that runner registers itself
	assert.Equal(t, &RunnerMeta{ID: wantID}, <-r.registered)
	// Terminate runner
	cancel()
	// Test that runner sends final status update
	assert.Equal(t, RunnerExited, <-updates)
}

type fakeRunnerClient struct {
	client

	registeredID resource.ID
	updates      chan RunnerStatus
}

func (f *fakeRunnerClient) register(ctx context.Context, opts registerOptions) (*RunnerMeta, error) {
	return &RunnerMeta{ID: f.registeredID}, nil
}

func (f *fakeRunnerClient) getJobs(ctx context.Context, agentID resource.ID) ([]*Job, error) {
	// block until context canceled
	select {
	case <-make(chan struct{}):
	case <-ctx.Done():
	}
	return nil, nil
}

func (f *fakeRunnerClient) updateStatus(ctx context.Context, agentID resource.ID, status RunnerStatus) error {
	f.updates <- status
	return nil
}

type fakeOperationSpawner struct {
	operationSpawner
}
