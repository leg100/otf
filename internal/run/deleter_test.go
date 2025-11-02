package run

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestDeleter(t *testing.T) {
	tests := []struct {
		name      string
		run       *Run
		threshold time.Duration
		// expect run to be deleted or not
		delete bool
	}{
		{
			name: "delete old run",
			run: &Run{
				// month-old
				CreatedAt: time.Now().Add(-time.Hour * 24 * 30),
			},
			threshold: time.Hour * 24,
			delete:    true,
		},
		{
			name: "don't delete newish run",
			run: &Run{
				// hour-old
				CreatedAt: time.Now().Add(-time.Hour),
			},
			threshold: time.Hour * 24,
			delete:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeDeleterRunClient{run: tt.run}
			deleter := &Deleter{
				Runs:         client,
				AgeThreshold: tt.threshold,
			}
			deleter.deleteRuns(context.Background())
			assert.Equal(t, tt.delete, client.deleted)
		})
	}
}

type fakeDeleterRunClient struct {
	run     *Run
	deleted bool
}

func (f *fakeDeleterRunClient) List(ctx context.Context, opts ListOptions) (*resource.Page[*Run], error) {
	if f.run.CreatedAt.Before(*opts.BeforeCreatedAt) {
		return resource.NewPage([]*Run{f.run}, resource.PageOptions{}, nil), nil
	}
	return resource.NewPage([]*Run{}, resource.PageOptions{}, nil), nil
}

func (f *fakeDeleterRunClient) Delete(ctx context.Context, runID resource.TfeID) error {
	f.deleted = true
	return nil
}
