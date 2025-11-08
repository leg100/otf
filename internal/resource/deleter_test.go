package resource

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDeleter(t *testing.T) {
	tests := []struct {
		name      string
		resource  *fakeDeleterResource
		threshold time.Duration
		// expect resource to be deleted or not
		delete bool
	}{
		{
			name: "delete old resource",
			resource: &fakeDeleterResource{
				// month-old
				CreatedAt: time.Now().Add(-time.Hour * 24 * 30),
			},
			threshold: time.Hour * 24,
			delete:    true,
		},
		{
			name: "don't delete newish resource",
			resource: &fakeDeleterResource{
				// hour-old
				CreatedAt: time.Now().Add(-time.Hour),
			},
			threshold: time.Hour * 24,
			delete:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeDeleterClient{res: tt.resource}
			deleter := &Deleter[*fakeDeleterResource]{
				Client:       client,
				AgeThreshold: tt.threshold,
			}
			deleter.deleteResources(context.Background())
			assert.Equal(t, tt.delete, client.deleted)
		})
	}
}

type fakeDeleterResource struct {
	ID        TfeID
	CreatedAt time.Time
}

// GetID implements resource.deleteableResource
func (r *fakeDeleterResource) GetID() TfeID { return r.ID }

type fakeDeleterClient struct {
	res     *fakeDeleterResource
	deleted bool
}

func (f *fakeDeleterClient) ListOlderThan(ctx context.Context, t time.Time) ([]*fakeDeleterResource, error) {
	if f.res.CreatedAt.Before(t) {
		return []*fakeDeleterResource{f.res}, nil
	}
	return nil, nil
}

func (f *fakeDeleterClient) Delete(context.Context, TfeID) error {
	f.deleted = true
	return nil
}
