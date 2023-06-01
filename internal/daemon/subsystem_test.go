package daemon

import (
	"context"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestSubsystem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		backoff   bool
		exclusive bool
	}{
		{"default", false, false},
		{"backoff", true, false},
		{"backoff and wait and lock", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &Subsystem{
				Name:           tt.name,
				System:         &fakeStartable{},
				Logger:         logr.Discard(),
				BackoffRestart: tt.backoff,
				Exclusive:      tt.exclusive,
			}
			if tt.exclusive {
				sub.DB = &fakeWaitAndLock{}
				sub.LockID = internal.Int64(123)
			}
			err := sub.Start(ctx, &errgroup.Group{})
			require.NoError(t, err)
		})
	}
}

type (
	fakeStartable   struct{}
	fakeWaitAndLock struct {
		internal.DB
	}
)

func (f *fakeStartable) Start(ctx context.Context) error {
	return nil
}

func (f *fakeWaitAndLock) WaitAndLock(ctx context.Context, id int64, fn func() error) error {
	return fn()
}
