package variable

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable_UpdateHandler(t *testing.T) {
	tests := []struct {
		name     string
		existing CreateVariableOptions
		updated  url.Values
		want     func(t *testing.T, got *Variable)
	}{
		{
			name: "overwrite everything",
			existing: CreateVariableOptions{
				Key:      internal.String("foo"),
				Value:    internal.String("bar"),
				Category: VariableCategoryPtr(CategoryTerraform),
			},
			updated: url.Values{
				"key":       {"new-key"},
				"value":     {"new-value"},
				"category":  {"env"},
				"sensitive": {"on"},
				"hcl":       {"on"},
			},
			want: func(t *testing.T, got *Variable) {
				assert.Equal(t, "new-key", got.Key)
				assert.Equal(t, "new-value", got.Value)
				assert.Equal(t, CategoryEnv, got.Category)
				assert.True(t, got.Sensitive)
				assert.True(t, got.HCL)
			},
		},
		{
			name: "update sensitive value",
			existing: CreateVariableOptions{
				Key:       internal.String("foo"),
				Value:     internal.String("topsecret"),
				Category:  VariableCategoryPtr(CategoryTerraform),
				Sensitive: internal.Bool(true),
			},
			updated: url.Values{
				"value": {"evenmoretopsecret"},
			},
			want: func(t *testing.T, got *Variable) {
				assert.Equal(t, "evenmoretopsecret", got.Value)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create existing variable for test to update
			v := newTestVariable(t, "ws-123", tt.existing)

			r := httptest.NewRequest("POST", "/?variable_id="+v.ID, strings.NewReader(tt.updated.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			fakeHTMLApp(t, v).update(w, r)

			if assert.Equal(t, 302, w.Code, "got body: %s", w.Body.String()) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, paths.Variables(v.WorkspaceID), redirect.Path)
			}
			tt.want(t, v)
		})
	}
}

func fakeHTMLApp(t *testing.T, variable *Variable) *web {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &web{
		Renderer: renderer,
		svc:      &fakeService{variable: variable},
	}
}
