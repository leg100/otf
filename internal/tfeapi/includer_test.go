package tfeapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncluder(t *testing.T) {
	type (
		bazType struct {
			ID string
		}
		barType struct {
			Baz bazType
			ID  string
		}
		fooType struct {
			Bar barType
			ID  string
		}
	)
	var (
		foo = &fooType{
			ID: "foo-id",
		}
		bar = &barType{
			ID: "bar-id",
		}
		baz = &bazType{
			ID: "baz-id",
		}
	)

	tests := []struct {
		name          string
		query         string
		resource      any
		registrations map[IncludeName]IncludeFunc
		want          []any
	}{
		{
			name:     "simple include",
			query:    "/foo?include=bar",
			resource: foo,
			registrations: map[IncludeName]IncludeFunc{
				IncludeName("bar"): func(_ context.Context, v any) ([]any, error) {
					assert.Equal(t, foo, v)
					return []any{bar}, nil
				},
			},
			want: []any{bar},
		},
		{
			name:     "multiple includes",
			query:    "/foo?include=bar,baz",
			resource: foo,
			registrations: map[IncludeName]IncludeFunc{
				IncludeName("bar"): func(_ context.Context, v any) ([]any, error) {
					assert.Equal(t, foo, v)
					return []any{bar}, nil
				},
				IncludeName("baz"): func(_ context.Context, v any) ([]any, error) {
					assert.Equal(t, foo, v)
					return []any{baz}, nil
				},
			},
			want: []any{bar, baz},
		},
		{
			name:     "include transitive relation",
			query:    "/foo?include=bar.baz",
			resource: foo,
			registrations: map[IncludeName]IncludeFunc{
				IncludeName("bar"): func(_ context.Context, v any) ([]any, error) {
					assert.Equal(t, foo, v)
					return []any{bar}, nil
				},
				IncludeName("baz"): func(_ context.Context, v any) ([]any, error) {
					assert.Equal(t, bar, v)
					return []any{baz}, nil
				},
			},
			want: []any{bar, baz},
		},
		{
			name:     "multiple resources",
			query:    "/foo?include=bar",
			resource: []any{fooType{ID: "foo-1"}, fooType{ID: "foo-2"}},
			registrations: map[IncludeName]IncludeFunc{
				IncludeName("bar"): func(_ context.Context, v any) ([]any, error) {
					return []any{bar}, nil
				},
			},
			want: []any{bar, bar},
		},
		{
			name:     "registered func returns nil",
			query:    "/?include=bar",
			resource: foo,
			registrations: map[IncludeName]IncludeFunc{
				IncludeName("bar"): func(_ context.Context, v any) ([]any, error) {
					return nil, nil
				},
			},
			want: nil,
		},
		{
			name:  "unregistered resource",
			query: "/?include=doesnotexist",
			want:  nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inc := &includer{registrations: tt.registrations}
			r := httptest.NewRequest("GET", tt.query, nil)
			got, err := inc.addIncludes(r, tt.resource)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}

}
