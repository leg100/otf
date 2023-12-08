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
				Key:             internal.String("foo"),
				Value:           internal.String("bar"),
				Category:        VariableCategoryPtr(CategoryTerraform),
				generateVersion: func() string { return "" },
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
			name: "disable hcl",
			existing: CreateVariableOptions{
				Key:             internal.String("foo"),
				Value:           internal.String("bar"),
				Category:        VariableCategoryPtr(CategoryTerraform),
				HCL:             internal.Bool(true),
				generateVersion: func() string { return "" },
			},
			// If the user unchecks the HCL checkbox then no form value is sent
			// but the handler should interpret the absence of the value as
			// 'false'.
			updated: url.Values{},
			want: func(t *testing.T, got *Variable) {
				assert.False(t, got.HCL)
			},
		},
		{
			name: "update sensitive value",
			existing: CreateVariableOptions{
				Key:             internal.String("foo"),
				Value:           internal.String("topsecret"),
				Category:        VariableCategoryPtr(CategoryTerraform),
				Sensitive:       internal.Bool(true),
				generateVersion: func() string { return "" },
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
			v, err := newVariable(nil, tt.existing)
			require.NoError(t, err)

			r := httptest.NewRequest("POST", "/?variable_id="+v.ID, strings.NewReader(tt.updated.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			fakeWebApp(t, "ws-123", v).updateWorkspaceVariable(w, r)

			if assert.Equal(t, 302, w.Code, "got body: %s", w.Body.String()) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, paths.Variables("ws-123"), redirect.Path)
			}
			tt.want(t, v)
		})
	}
}

func fakeWebApp(t *testing.T, workspaceID string, v *Variable) *web {
	renderer, err := html.NewRenderer(false)
	require.NoError(t, err)
	return &web{
		Renderer:  renderer,
		variables: &fakeService{v: v, workspaceID: workspaceID},
	}
}
