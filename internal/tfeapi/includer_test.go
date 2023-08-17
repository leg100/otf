package tfeapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncluder(t *testing.T) {
	type bar struct {
		ID string
	}
	type foo struct {
		Bar bar
	}
	inc := &includer{
		registrations: map[IncludeName]IncludeFunc{
			IncludeName("bar"): func(ctx context.Context, v any) (any, error) {
				return &bar{ID: "bar-id"}, nil
			},
		},
	}
	r := httptest.NewRequest("GET", "/?include=bar", nil)
	got, err := inc.addIncludes(r, &foo{})
	require.NoError(t, err)
	assert.Equal(t, []any{&bar{ID: "bar-id"}}, got)
}
