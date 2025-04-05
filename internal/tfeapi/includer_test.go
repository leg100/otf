package tfeapi

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIncluder(t *testing.T) {
	type (
		baz struct {
			ID resource.TfeID
		}
		bar struct {
			Baz baz
			ID  resource.TfeID
		}
		foo struct {
			Bar bar
			ID  resource.TfeID
		}
	)
	fooResource := foo{ID: resource.NewTfeID("foo")}
	fooResource2 := foo{ID: resource.NewTfeID("foo")}
	barResource := bar{ID: resource.NewTfeID("bar")}
	barResource2 := bar{ID: resource.NewTfeID("bar")}
	bazResource := baz{ID: resource.NewTfeID("baz")}

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
			resource: &fooResource,
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &fooResource, v)
						return []any{&barResource}, nil
					},
				},
			},
			want: []any{&barResource},
		},
		{
			name:     "multiple includes",
			query:    "/foo?include=bar,baz",
			resource: &fooResource,
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &fooResource, v)
						return []any{&barResource}, nil
					},
				},
				IncludeName("baz"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &fooResource, v)
						return []any{&bazResource}, nil
					},
				},
			},
			want: []any{&barResource, &bazResource},
		},
		{
			name:     "include transitive relation",
			query:    "/foo?include=bar.baz",
			resource: &fooResource,
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &fooResource, v)
						return []any{&barResource}, nil
					},
				},
				IncludeName("baz"): {
					func(_ context.Context, v any) ([]any, error) {
						assert.Equal(t, &barResource, v)
						return []any{&bazResource}, nil
					},
				},
			},
			want: []any{&barResource, &bazResource},
		},
		{
			name:     "multiple resources",
			query:    "/foo?include=bar",
			resource: []any{fooResource, fooResource2},
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						return []any{&barResource}, nil
					},
				},
			},
			want: []any{&barResource, &barResource},
		},
		{
			name:     "multiple registrations for same include",
			query:    "/foo?include=bar",
			resource: &fooResource,
			registrations: map[IncludeName][]IncludeFunc{
				IncludeName("bar"): {
					func(_ context.Context, v any) ([]any, error) {
						return []any{&barResource}, nil
					},
					func(_ context.Context, v any) ([]any, error) {
						return []any{&barResource2}, nil
					},
				},
			},
			want: []any{&barResource, &barResource2},
		},
		{
			name:     "registered func returns nil",
			query:    "/?include=bar",
			resource: &fooResource,
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

//func TestGetRelationshipID(t *testing.T) {
//	type (
//		bar struct {
//			ID resource.TfeID `jsonapi:"primary,bars"`
//		}
//		foo struct {
//			ID  resource.TfeID `jsonapi:"primary,foos"`
//			Bar *bar           `jsonapi:"relationship" json:"bar"`
//		}
//	)
//}
