package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/jsonapi"
	"github.com/leg100/go-tfe"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	s := &Server{}
	s.WorkspaceService = &mock.WorkspaceService{
		CreateWorkspaceFn: func(org string, opts *tfe.WorkspaceCreateOptions) (*tfe.Workspace, error) {
			return mock.NewWorkspace(*opts.Name, "ws-123", org), nil
		},
		UpdateWorkspaceFn: func(name, org string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
			return mock.NewWorkspace(*opts.Name, "ws-123", org), nil
		},
		UpdateWorkspaceByIDFn: func(id string, opts *tfe.WorkspaceUpdateOptions) (*tfe.Workspace, error) {
			return mock.NewWorkspace(*opts.Name, "ws-123", "automatize"), nil
		},
		GetWorkspaceFn: func(name, org string) (*tfe.Workspace, error) {
			return mock.NewWorkspace(name, "ws-123", org), nil
		},
		GetWorkspaceByIDFn: func(id string) (*tfe.Workspace, error) {
			return mock.NewWorkspace("dev", "ws-123", "automatize"), nil
		},
		ListWorkspaceFn: func(org string, opts tfe.WorkspaceListOptions) (*tfe.WorkspaceList, error) {
			return mock.NewWorkspaceList("dev", "ws-123", org, opts), nil
		},
		DeleteWorkspaceFn: func(name, org string) error {
			return nil
		},
		DeleteWorkspaceByIDFn: func(id string) error {
			return nil
		},
	}

	tests := []struct {
		name       string
		method     string
		url        string
		payload    map[string]interface{}
		wantStatus int
		wantResp   map[string]interface{}
	}{
		{
			name:       "get",
			method:     "GET",
			url:        "/api/v2/organizations/automatize/workspaces/dev",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "dev",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
			},
		},
		{
			name:       "get by id",
			method:     "GET",
			url:        "/api/v2/workspaces/ws-123",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "dev",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
			},
		},
		{
			name:       "get with included org",
			method:     "GET",
			url:        "/api/v2/workspaces/ws-123?include=organizations",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "dev",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
				"included": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"collaborator-auth-policy": "",
							"cost-estimation-enabled":  false,
							"email":                    "",
							"external-id":              "",
							"owners-team-saml-role-id": "",
							"permissions":              interface{}(nil),
							"saml-enabled":             false,
							"session-remember":         float64(0),
							"session-timeout":          float64(0),
							"two-factor-conformant":    false,
						},
						"id":   "automatize",
						"type": "organizations",
					},
				},
			},
		},
		{
			name:       "list",
			method:     "GET",
			url:        "/api/v2/organizations/automatize/workspaces",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"name": "dev",
							"actions": map[string]interface{}{
								"is-destroyable": false,
							},
							"agent-pool-id":          "",
							"allow-destroy-plan":     false,
							"apply-duration-average": float64(0),
							"auto-apply":             false,
							"can-queue-destroy-plan": false,
							"description":            "",
							"environment":            "",
							"execution-mode":         "",
							"file-triggers-enabled":  false,
							"global-remote-state":    false,
							"locked":                 false,
							"migration-environment":  "",
							"operations":             false,
							"permissions": map[string]interface{}{
								"can-destroy":         false,
								"can-force-unlock":    false,
								"can-lock":            false,
								"can-queue-apply":     false,
								"can-queue-destroy":   false,
								"can-queue-run":       false,
								"can-read-settings":   false,
								"can-unlock":          false,
								"can-update":          false,
								"can-update-variable": false,
							},
							"plan-duration-average":         float64(0),
							"policy-check-failures":         float64(0),
							"queue-all-runs":                false,
							"resource-count":                float64(0),
							"run-failures":                  float64(0),
							"source-name":                   "",
							"source-url":                    "",
							"speculative-enabled":           false,
							"structured-run-output-enabled": false,
							"terraform-version":             "",
							"trigger-prefixes":              []interface{}{},
							"vcs-repo": map[string]interface{}{"branch": "",
								"display-identifier":  "",
								"identifier":          "",
								"ingress-submodules":  false,
								"oauth-token-id":      "",
								"repository-http-url": "",
								"service-provider":    "",
							},
							"working-directory":         "",
							"workspace-kpis-runs-count": float64(0),
						},
						"id": "ws-123",
						"relationships": map[string]interface{}{
							"agent-pool": map[string]interface{}{
								"data": interface{}(nil),
							},
							"current-run": map[string]interface{}{
								"data": interface{}(nil),
							},
							"organization": map[string]interface{}{
								"data": map[string]interface{}{
									"id":   "automatize",
									"type": "organizations",
								},
							},
							"ssh-key": map[string]interface{}{
								"data": interface{}(nil),
							},
						},
						"type": "workspaces",
					},
				},
				"meta": map[string]interface{}{
					"pagination": map[string]interface{}{
						"prev-page":    float64(1),
						"current-page": float64(1),
						"next-page":    float64(1),
						"total-count":  float64(1),
						"total-pages":  float64(1),
					},
				},
			},
		},
		{
			name:   "create",
			method: "POST",
			url:    "/api/v2/organizations/automatize/workspaces",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "workspaces",
					"attributes": map[string]interface{}{
						"name": "dev",
					},
				},
			},
			wantStatus: 201,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "dev",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
			},
		},
		{
			name:   "update",
			method: "PATCH",
			url:    "/api/v2/organizations/automatize/workspaces/dev",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "workspaces",
					"attributes": map[string]interface{}{
						"name": "staging",
					},
				},
			},
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "staging",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
			},
		},
		{
			name:   "update by id",
			method: "PATCH",
			url:    "/api/v2/workspaces/ws-123",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "workspaces",
					"attributes": map[string]interface{}{
						"name": "staging",
					},
				},
			},
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"name": "staging",
						"actions": map[string]interface{}{
							"is-destroyable": false,
						},
						"agent-pool-id":          "",
						"allow-destroy-plan":     false,
						"apply-duration-average": float64(0),
						"auto-apply":             false,
						"can-queue-destroy-plan": false,
						"description":            "",
						"environment":            "",
						"execution-mode":         "",
						"file-triggers-enabled":  false,
						"global-remote-state":    false,
						"locked":                 false,
						"migration-environment":  "",
						"operations":             false,
						"permissions": map[string]interface{}{
							"can-destroy":         false,
							"can-force-unlock":    false,
							"can-lock":            false,
							"can-queue-apply":     false,
							"can-queue-destroy":   false,
							"can-queue-run":       false,
							"can-read-settings":   false,
							"can-unlock":          false,
							"can-update":          false,
							"can-update-variable": false,
						},
						"plan-duration-average":         float64(0),
						"policy-check-failures":         float64(0),
						"queue-all-runs":                false,
						"resource-count":                float64(0),
						"run-failures":                  float64(0),
						"source-name":                   "",
						"source-url":                    "",
						"speculative-enabled":           false,
						"structured-run-output-enabled": false,
						"terraform-version":             "",
						"trigger-prefixes":              []interface{}{},
						"vcs-repo": map[string]interface{}{"branch": "",
							"display-identifier":  "",
							"identifier":          "",
							"ingress-submodules":  false,
							"oauth-token-id":      "",
							"repository-http-url": "",
							"service-provider":    "",
						},
						"working-directory":         "",
						"workspace-kpis-runs-count": float64(0),
					},
					"id": "ws-123",
					"relationships": map[string]interface{}{
						"agent-pool": map[string]interface{}{
							"data": interface{}(nil),
						},
						"current-run": map[string]interface{}{
							"data": interface{}(nil),
						},
						"organization": map[string]interface{}{
							"data": map[string]interface{}{
								"id":   "automatize",
								"type": "organizations",
							},
						},
						"ssh-key": map[string]interface{}{
							"data": interface{}(nil),
						},
					},
					"type": "workspaces",
				},
			},
		},
		{
			name:       "delete",
			method:     "DELETE",
			url:        "/api/v2/organizations/automatize/workspaces/dev",
			wantStatus: 204,
		},
		{
			name:       "delete by id",
			method:     "DELETE",
			url:        "/api/v2/workspaces/ws-123",
			wantStatus: 204,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Construct http req
			buf := new(bytes.Buffer)
			if tt.payload != nil {
				body, err := json.Marshal(tt.payload)
				require.NoError(t, err)
				buf = bytes.NewBuffer(body)
			}

			req, err := http.NewRequest(tt.method, tt.url, buf)
			require.NoError(t, err)
			req.Header.Set("Accept", jsonapi.MediaType)

			rr := httptest.NewRecorder()

			// Dispatch request
			router := NewRouter(s)
			router.ServeHTTP(rr, req)

			// Assertions
			assert.Equal(t, tt.wantStatus, rr.Code)

			var got map[string]interface{}
			if rr.Body.Len() > 0 {
				got = make(map[string]interface{})
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
			}
			assert.Equal(t, tt.wantResp, got)
		})
	}
}
