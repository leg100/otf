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
		baz struct {
			ID string
		}
		bar struct {
			Baz baz
			ID  string
		}
		foo struct {
			Bar bar
			ID  string
		}
	)
	tests := []struct {
		name          string
		query         string
		resource      any
		registrations map[IncludeName][]IncludeFunc
		want          []any
	}{
		{
			name:     "simple include",
			query:    "/foo?include=bar",
			resource: &foo{ID: "foo-1"},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &foo{ID: "foo-1"}, v)
						return []any{&bar{ID: "bar-1"}}, nil
					},
				},
			},
			want: []any{&bar{ID: "bar-1"}},
		},
		{
			name:     "multiple includes",
			query:    "/foo?include=bar,baz",
			resource: &foo{ID: "foo-1"},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &foo{ID: "foo-1"}, v)
						return []any{&bar{ID: "bar-1"}}, nil
					},
				},
				IncludeName("baz"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &foo{ID: "foo-1"}, v)
						return []any{&baz{"baz-1"}}, nil
					},
				},
			},
			want: []any{&bar{ID: "bar-1"}, &baz{"baz-1"}},
		},
		{
			name:     "include transitive relation",
			query:    "/foo?include=bar.baz",
			resource: &foo{ID: "foo-1"},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &foo{ID: "foo-1"}, v)
						return []any{&bar{ID: "bar-1"}}, nil
					},
				},
				IncludeName("baz"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &bar{ID: "bar-1"}, v)
						return []any{&baz{"baz-1"}}, nil
					},
				},
			},
			want: []any{&bar{ID: "bar-1"}, &baz{"baz-1"}},
		},
		{
			name:     "multiple resources",
			query:    "/foo?include=bar",
			resource: []any{foo{ID: "foo-1"}, foo{ID: "foo-2"}},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						return []any{&bar{ID: "bar-1"}}, nil
					},
				},
			},
			want: []any{&bar{ID: "bar-1"}, &bar{ID: "bar-1"}},
		},
		{
			name:     "multiple registrations for same include",
			query:    "/foo?include=bar",
			resource: &foo{ID: "foo-1"},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						return []any{&bar{ID: "bar-1"}}, nil
					},
					func(_ context.Context, v any) ([]any, error) {
						return []any{&bar{ID: "bar-2"}}, nil
					},
				},
			},
			want: []any{&bar{ID: "bar-1"}, &bar{ID: "bar-2"}},
		},
		{
			name:     "registered func returns nil",
			query:    "/?include=bar",
			resource: &foo{ID: "foo-1"},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						return nil, nil
					},
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
