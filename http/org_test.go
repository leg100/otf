package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/jsonapi"
	"github.com/hashicorp/go-tfe"
	"github.com/leg100/ots"
	"github.com/leg100/ots/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganization(t *testing.T) {
	s := &Server{}
	s.OrganizationService = &mock.OrganizationService{
		CreateOrganizationFn: func(name string, opts *tfe.OrganizationCreateOptions) (*ots.Organization, error) {
			createdAt, err := time.Parse(time.RFC3339, "2021-06-07T08:23:36Z")
			if err != nil {
				return nil, err
			}
			return ots.NewOrganizationFromOptions(opts, ots.OrgCreatedAt(createdAt))

		},
		UpdateOrganizationFn: func(name string, opts *tfe.OrganizationUpdateOptions) (*ots.Organization, error) {
			org := &ots.Organization{
				Name:  "automatize",
				Email: "leg100",
			}
			if err := ots.UpdateOrganizationFromOptions(org, opts); err != nil {
				return nil, err
			}
			return org, nil

		},
		GetOrganizationFn: func(name string) (*ots.Organization, error) {
			if name != "automatize" {
				return nil, errors.New("not found")
			}
			return &ots.Organization{
				Name:  "automatize",
				Email: "leg100",
			}, nil
		},
		ListOrganizationFn: func() ([]*ots.Organization, error) {
			return []*ots.Organization{
				{
					Name:  "automatize",
					Email: "leg100",
				},
			}, nil
		},
		DeleteOrganizationFn: func(name string) error {
			if name != "automatize" {
				return errors.New("not found")
			}
			return nil
		},
		GetEntitlementsFn: func(name string) (*ots.Entitlements, error) {
			if name != "automatize" {
				return nil, errors.New("not found")
			}
			return ots.NewEntitlements(name), nil
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
			path:       "/organizations/automatize",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"collaborator-auth-policy": "",
						"cost-estimation-enabled":  false,
						"email":                    "leg100",
						"enterprise-plan":          "",
						"external-id":              "",
						"owners-team-saml-role-id": "",
						"permissions":              interface{}(nil),
						"saml-enabled":             false,
						"session-remember":         float64(0),
						"session-timeout":          float64(0),
						"two-factor-conformant":    false,
					},
					"id": "automatize",
					"links": map[string]interface{}{
						"self": "/v2/api/organizations/automatize",
					},
					"type": "organizations",
				},
			},
		},
		{
			name:       "list",
			method:     "GET",
			path:       "/organizations",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"attributes": map[string]interface{}{
							"collaborator-auth-policy": "",
							"cost-estimation-enabled":  false,
							"email":                    "leg100",
							"enterprise-plan":          "",
							"external-id":              "",
							"owners-team-saml-role-id": "",
							"permissions":              interface{}(nil),
							"saml-enabled":             false,
							"session-remember":         float64(0),
							"session-timeout":          float64(0),
							"two-factor-conformant":    false,
						},
						"id": "automatize",
						"links": map[string]interface{}{
							"self": "/v2/api/organizations/automatize",
						},
						"type": "organizations",
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
			path:       "/organizations/automatize",
			wantStatus: 201,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"collaborator-auth-policy": "password",
						"cost-estimation-enabled":  true,
						"created-at":               "2021-06-07T08:23:36Z",
						"email":                    "leg100@automatize.co.uk",
						"enterprise-plan":          "",
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
					"id": "automatize",
					"links": map[string]interface{}{
						"self": "/v2/api/organizations/automatize",
					},
					"type": "organizations",
				},
			},
		},
		{
			name:   "update",
			method: "PATCH",
			path:   "/organizations/automatize",
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
						"collaborator-auth-policy": "",
						"cost-estimation-enabled":  false,
						"email":                    "leg101@automatize.co.uk",
						"enterprise-plan":          "",
						"external-id":              "",
						"owners-team-saml-role-id": "",
						"permissions":              interface{}(nil),
						"saml-enabled":             false,
						"session-remember":         float64(0),
						"session-timeout":          float64(0),
						"two-factor-conformant":    false,
					},
					"id": "automatize",
					"links": map[string]interface{}{
						"self": "/v2/api/organizations/automatize",
					},
					"type": "organizations",
				},
			},
		},
		{
			name:       "delete",
			method:     "DELETE",
			path:       "/organizations/automatize",
			wantStatus: 204,
		},
		{
			name:       "get entitlements",
			method:     "GET",
			path:       "/organizations/automatize/entitlement-set",
			wantStatus: 200,
			wantResp: map[string]interface{}{
				"data": map[string]interface{}{
					"attributes": map[string]interface{}{
						"agents":                  false,
						"audit-logging":           false,
						"cost-estimation":         false,
						"operations":              false,
						"private-module-registry": false,
						"sentinel":                false,
						"sso":                     false,
						"state-storage":           false,
						"teams":                   false,
						"vcs-integrations":        false,
					},
					"id": "automatize",
					"links": map[string]interface{}{
						"self": "/v2/api/entitlement-set/automatize",
					},
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
