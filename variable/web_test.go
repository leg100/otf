package variable

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/html"
	"github.com/leg100/otf/http/html/paths"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable_Update(t *testing.T) {
	org := otf.NewTestOrganization(t)
	ws := otf.NewTestWorkspace(t, org)

	tests := []struct {
		name     string
		existing otf.CreateVariableOptions
		form     url.Values
		want     func(t *testing.T, got *Variable)
	}{
		{
			name: "overwrite everything",
			existing: otf.CreateVariableOptions{
				Key:      otf.String("foo"),
				Value:    otf.String("bar"),
				Category: VariableCategoryPtr(otf.CategoryTerraform),
			},
			form: url.Values{
				"key":       {"new-key"},
				"value":     {"new-value"},
				"category":  {"env"},
				"sensitive": {"on"},
				"hcl":       {"on"},
			},
			want: func(t *testing.T, got *Variable) {
				assert.Equal(t, "new-key", got.Key())
				assert.Equal(t, "new-value", got.Value())
				assert.Equal(t, otf.CategoryEnv, got.Category())
				assert.True(t, got.Sensitive())
				assert.True(t, got.HCL())
			},
		},
		{
			name: "skip sensitive variable empty value",
			existing: otf.CreateVariableOptions{
				Key:      otf.String("foo"),
				Value:    otf.String("bar"),
				Category: VariableCategoryPtr(otf.CategoryTerraform),
			},
			form: url.Values{
				"key":       {"foo"},
				"value":     {""},
				"sensitive": {"on"},
				"category":  {"terraform"},
			},
			want: func(t *testing.T, got *Variable) {
				// should get original value
				assert.Equal(t, "bar", got.Value())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create existing variable for test to update
			v := NewTestVariable(t, ws, tt.existing)

			r := httptest.NewRequest("POST", "/?variable_id="+v.ID, strings.NewReader(tt.form.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			fakeHTMLApp(t, v).update(w, r)

			if assert.Equal(t, 302, w.Code) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, paths.Variables(v.WorkspaceID), redirect.Path)
			}
			tt.want(t, v)
		})
	}
}

func fakeHTMLApp(t *testing.T, variable *Variable) *web {
	renderer, err := html.NewViewEngine(false)
	require.NoError(t, err)
	return &web{
		Renderer: renderer,
		app:      &fakeService{variable: variable},
	}
}
