package daemon

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

// fakeStartable blocks until context is cancelled so the backoff retry loop
// doesn't spin.
type fakeStartable struct {
	startCalled atomic.Bool
}

func (f *fakeStartable) Start(ctx context.Context) error {
	f.startCalled.Store(true)
	<-ctx.Done()
	return nil
}

// fakeStartableWithSignal closes its started channel as soon as Start is
// invoked, satisfying the Started() wait in startSubsystems.
type fakeStartableWithSignal struct {
	startedCh chan struct{}
}

func (f *fakeStartableWithSignal) Start(ctx context.Context) error {
	// Signal that the subsystem has started.
	select {
	case <-f.startedCh:
		// already closed (retry path); do nothing
	default:
		close(f.startedCh)
	}
	<-ctx.Done()
	return nil
}

func (f *fakeStartableWithSignal) Started() <-chan struct{} {
	return f.startedCh
}

// neverStartedStartable implements Started() but never closes the channel.
type neverStartedStartable struct {
	startedCh chan struct{}
}

func (n *neverStartedStartable) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (n *neverStartedStartable) Started() <-chan struct{} {
	return n.startedCh
}

// fakeLocker grants the lock immediately by calling fn directly.
type fakeLocker struct{}

func (f *fakeLocker) WaitForExclusiveLock(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// blockingLocker never grants the lock; it blocks until ctx is cancelled.
type blockingLocker struct{}

func (b *blockingLocker) WaitForExclusiveLock(ctx context.Context, fn func(context.Context) error) error {
	<-ctx.Done()
	return ctx.Err()
}

// errorLocker returns an error without ever calling fn.
type errorLocker struct {
	err error
}

func (e *errorLocker) WaitForExclusiveLock(_ context.Context, _ func(context.Context) error) error {
	return e.err
}

func makeSubsystem(name string, sys Startable, exclusive bool) *Subsystem {
	return &Subsystem{
		Name:      name,
		Logger:    logr.Discard(),
		System:    sys,
		Exclusive: exclusive,
	}
}

// TestStartSubsystems_NonExclusive checks that a non-exclusive subsystem is
// started when startSubsystems is called.
func TestStartSubsystems_NonExclusive(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, ctx := errgroup.WithContext(ctx)

	sys := &fakeStartable{}
	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("non-exclusive", sys, false)}, &fakeLocker{})
	require.NoError(t, err)

	cancel()
	require.NoError(t, g.Wait())
	assert.True(t, sys.startCalled.Load(), "expected non-exclusive subsystem Start to be called")
}

// TestStartSubsystems_Exclusive checks that an exclusive subsystem is started
// once the locker grants the lock.
func TestStartSubsystems_Exclusive(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, ctx := errgroup.WithContext(ctx)

	sys := &fakeStartable{}
	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("exclusive", sys, true)}, &fakeLocker{})
	require.NoError(t, err)

	cancel()
	require.NoError(t, g.Wait())
	assert.True(t, sys.startCalled.Load(), "expected exclusive subsystem Start to be called after lock was granted")
}

// TestStartSubsystems_ExclusiveNotStartedWithoutLock checks that an exclusive
// subsystem is never started when the locker blocks indefinitely.
func TestStartSubsystems_ExclusiveNotStartedWithoutLock(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, ctx := errgroup.WithContext(ctx)

	sys := &fakeStartable{}
	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("exclusive", sys, true)}, &blockingLocker{})
	require.NoError(t, err)

	// Cancel before the lock is ever granted.
	cancel()
	require.NoError(t, g.Wait())
	assert.False(t, sys.startCalled.Load(), "exclusive subsystem should not have been started without the lock")
}

// TestStartSubsystems_WaitsForStartedSignal checks that startSubsystems blocks
// until a non-exclusive subsystem closes its Started() channel.
func TestStartSubsystems_WaitsForStartedSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	sys := &fakeStartableWithSignal{startedCh: make(chan struct{})}
	// startSubsystems must not return until Started() fires; the fake closes
	// the channel immediately inside its Start method.
	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("with-signal", sys, false)}, &fakeLocker{})
	require.NoError(t, err)

	cancel()
	require.NoError(t, g.Wait())
}

// TestStartSubsystems_ContextCancelledWaitingForStarted checks that if the
// context is cancelled while startSubsystems is waiting for a Started()
// channel, it returns context.Canceled.
func TestStartSubsystems_ContextCancelledWaitingForStarted(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, ctx := errgroup.WithContext(ctx)

	sys := &neverStartedStartable{startedCh: make(chan struct{})}
	// Cancel the context so the Started() wait fires ctx.Done() immediately.
	cancel()

	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("never-started", sys, false)}, &fakeLocker{})
	require.ErrorIs(t, err, context.Canceled)
}

// TestStartSubsystems_LockerError checks that a locker error is propagated via
// the errgroup when the context has not been cancelled.
func TestStartSubsystems_LockerError(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	lockerErr := errors.New("advisory lock failed")
	err := startSubsystems(ctx, logr.Discard(), g, []*Subsystem{makeSubsystem("exclusive", &fakeStartable{}, true)}, &errorLocker{err: lockerErr})
	require.NoError(t, err, "startSubsystems itself should not error")

	err = g.Wait()
	require.ErrorIs(t, err, lockerErr, "locker error should be returned by errgroup")
}

// TestStartSubsystems_Mixed checks that both exclusive and non-exclusive
// subsystems are started when given a permissive locker.
func TestStartSubsystems_Mixed(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	g, ctx := errgroup.WithContext(ctx)

	nonExclusive := &fakeStartable{}
	exclusive := &fakeStartable{}

	subsystems := []*Subsystem{
		makeSubsystem("non-exclusive", nonExclusive, false),
		makeSubsystem("exclusive", exclusive, true),
	}

	err := startSubsystems(ctx, logr.Discard(), g, subsystems, &fakeLocker{})
	require.NoError(t, err)

	cancel()
	require.NoError(t, g.Wait())

	assert.True(t, nonExclusive.startCalled.Load(), "non-exclusive subsystem should be started")
	assert.True(t, exclusive.startCalled.Load(), "exclusive subsystem should be started")
}
