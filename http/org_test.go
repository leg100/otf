package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/leg100/go-tfe"
	"github.com/leg100/jsonapi"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization(t *testing.T) {
	s := &Server{
		Logger: logr.Discard(),
	}
	s.OrganizationService = &mock.OrganizationService{
		CreateOrganizationFn: func(opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
			return mock.NewOrganization(*opts.Name, *opts.Email), nil

		},
		UpdateOrganizationFn: func(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
			return mock.NewOrganization(*opts.Name, *opts.Email), nil
		},
		GetOrganizationFn: func(name string) (*ots.Organization, error) {
			return mock.NewOrganization(name, "leg100@automatize.co.uk"), nil
		},
		ListOrganizationFn: func(opts tfe.OrganizationListOptions) (*ots.OrganizationList, error) {
			return mock.NewOrganizationList("automatize", "leg100@automatize.co.uk", opts), nil
		},
		DeleteOrganizationFn: func(name string) error {
			return nil
		},
		GetEntitlementsFn: func(name string) (*ots.Entitlements, error) {
			return ots.DefaultEntitlements("org-123"), nil
		},
	}

	tests := []struct {
		name       string
		method     string
		path       string
		payload    map[string]interface{}
		wantStatus int
		wantResp   map[string]interface{}
	}{
		{
			name:       "get",
			method:     "GET",
			path:       "/api/v2/organizations/automatize",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"collaborator-auth-policy": "password",
						"cost-estimation-enabled":  true,
						"email":                    "leg100@automatize.co.uk",
						"external-id":              "",
						"owners-team-saml-role-id": "",
						"permissions": map[string]interface{}{
							"can-create-team":                false,
							"can-create-workspace":           false,
							"can-create-workspace-migration": false,
							"can-destroy":                    false,
							"can-traverse":                   false,
							"can-update":                     false,
							"can-update-api-token":           false,
							"can-update-oauth":               false,
							"can-update-sentinel":            false,
						},
						"saml-enabled":          false,
						"session-remember":      float64(20160),
						"session-timeout":       float64(20160),
						"two-factor-conformant": false,
					},
					"id":   "automatize",
					"type": "organizations",
				},
			},
		},
		{
			name:       "list",
			method:     "GET",
			path:       "/api/v2/organizations",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"collaborator-auth-policy": "password",
							"cost-estimation-enabled":  true,
							"email":                    "leg100@automatize.co.uk",
							"external-id":              "",
							"owners-team-saml-role-id": "",
							"permissions": map[string]interface{}{
								"can-create-team":                false,
								"can-create-workspace":           false,
								"can-create-workspace-migration": false,
								"can-destroy":                    false,
								"can-traverse":                   false,
								"can-update":                     false,
								"can-update-api-token":           false,
								"can-update-oauth":               false,
								"can-update-sentinel":            false,
							},
							"saml-enabled":          false,
							"session-remember":      float64(20160),
							"session-timeout":       float64(20160),
							"two-factor-conformant": false,
						},
						"id":   "automatize",
						"type": "organizations",
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
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "organizations",
					"attributes": map[string]interface{}{
						"name":  "automatize",
						"email": "leg100@automatize.co.uk",
					},
				},
			},
			path:       "/api/v2/organizations",
			wantStatus: 201,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"collaborator-auth-policy": "password",
						"cost-estimation-enabled":  true,
						"email":                    "leg100@automatize.co.uk",
						"external-id":              "",
						"owners-team-saml-role-id": "",
						"permissions": map[string]interface{}{
							"can-create-team":                false,
							"can-create-workspace":           false,
							"can-create-workspace-migration": false,
							"can-destroy":                    false,
							"can-traverse":                   false,
							"can-update":                     false,
							"can-update-api-token":           false,
							"can-update-oauth":               false,
							"can-update-sentinel":            false,
						},
						"saml-enabled":          false,
						"session-remember":      float64(20160),
						"session-timeout":       float64(20160),
						"two-factor-conformant": false,
					},
					"id":   "automatize",
					"type": "organizations",
				},
			},
		},
		{
			name:   "update",
			method: "PATCH",
			path:   "/api/v2/organizations/automatize",
			payload: map[string]interface{}{
				"data": map[string]interface{}{
					"type": "organizations",
					"attributes": map[string]interface{}{
						"name":  "automatize",
						"email": "leg101@automatize.co.uk",
					},
				},
			},
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"collaborator-auth-policy": "password",
						"cost-estimation-enabled":  true,
						"email":                    "leg101@automatize.co.uk",
						"external-id":              "",
						"owners-team-saml-role-id": "",
						"permissions": map[string]interface{}{
							"can-create-team":                false,
							"can-create-workspace":           false,
							"can-create-workspace-migration": false,
							"can-destroy":                    false,
							"can-traverse":                   false,
							"can-update":                     false,
							"can-update-api-token":           false,
							"can-update-oauth":               false,
							"can-update-sentinel":            false,
						},
						"saml-enabled":          false,
						"session-remember":      float64(20160),
						"session-timeout":       float64(20160),
						"two-factor-conformant": false,
					},
					"id":   "automatize",
					"type": "organizations",
				},
			},
		},
		{
			name:       "delete",
			method:     "DELETE",
			path:       "/api/v2/organizations/automatize",
			wantStatus: 204,
		},
		{
			name:       "get entitlements",
			method:     "GET",
			path:       "/api/v2/organizations/automatize/entitlement-set",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"agents":                  false,
						"audit-logging":           false,
						"cost-estimation":         false,
						"operations":              true,
						"private-module-registry": false,
						"sentinel":                false,
						"sso":                     false,
						"state-storage":           true,
						"teams":                   false,
						"vcs-integrations":        false,
					},
					"id":   "org-123",
					"type": "entitlement-sets",
				},
			},
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

			req, err := http.NewRequest(tt.method, tt.path, buf)
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
