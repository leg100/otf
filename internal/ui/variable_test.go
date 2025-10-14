package ui

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html/paths"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/testutils"
	"github.com/leg100/otf/internal/variable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariable_UpdateHandler(t *testing.T) {
	tests := []struct {
		name     string
		existing variable.CreateVariableOptions
		updated  url.Values
		want     func(t *testing.T, got *variable.Variable)
	}{
		{
			name: "overwrite everything",
			existing: variable.CreateVariableOptions{
				Key:             internal.Ptr("foo"),
				Value:           internal.Ptr("bar"),
				Category:        internal.Ptr(variable.CategoryTerraform),
				generateVersion: func() string { return "" },
			},
			updated: url.Values{
				"key":       {"new-key"},
				"value":     {"new-value"},
				"category":  {"env"},
				"sensitive": {"on"},
				"hcl":       {"on"},
			},
			want: func(t *testing.T, got *variable.Variable) {
				assert.Equal(t, "new-key", got.Key)
				assert.Equal(t, "new-value", got.Value)
				assert.Equal(t, variable.CategoryEnv, got.Category)
				assert.True(t, got.Sensitive)
				assert.True(t, got.HCL)
			},
		},
		{
			name: "disable hcl",
			existing: variable.CreateVariableOptions{
				Key:             internal.Ptr("foo"),
				Value:           internal.Ptr("bar"),
				Category:        internal.Ptr(variable.CategoryTerraform),
				HCL:             internal.Ptr(true),
				generateVersion: func() string { return "" },
			},
			// If the user unchecks the HCL checkbox then no form value is sent
			// but the handler should interpret the absence of the value as
			// 'false'.
			updated: url.Values{},
			want: func(t *testing.T, got *variable.Variable) {
				assert.False(t, got.HCL)
			},
		},
		{
			name: "update sensitive value",
			existing: variable.CreateVariableOptions{
				Key:             internal.Ptr("foo"),
				Value:           internal.Ptr("topsecret"),
				Category:        internal.Ptr(variable.CategoryTerraform),
				Sensitive:       internal.Ptr(true),
				generateVersion: func() string { return "" },
			},
			updated: url.Values{
				"value": {"evenmoretopsecret"},
			},
			want: func(t *testing.T, got *variable.Variable) {
				assert.Equal(t, "evenmoretopsecret", got.Value)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// create existing variable for test to update
			v, err := newVariable(nil, tt.existing)
			require.NoError(t, err)

			r := httptest.NewRequest("POST", "/?variable_id="+v.ID.String(), strings.NewReader(tt.updated.Encode()))
			r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			workspaceID := testutils.ParseID(t, "ws-123")
			fakeVariableService(t, workspaceID, v).updateWorkspaceVariable(w, r)

			if assert.Equal(t, 302, w.Code, "got body: %s", w.Body.String()) {
				redirect, err := w.Result().Location()
				require.NoError(t, err)
				assert.Equal(t, paths.Variables(workspaceID), redirect.Path)
			}
			tt.want(t, v)
		})
	}
}

func fakeVariableService(t *testing.T, workspaceID resource.TfeID, v *variable.Variable) *variableHandlers {
	return &variableHandlers{
		variables: &fakeGithubService{v: v, workspaceID: workspaceID},
	}
}
