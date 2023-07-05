package api

import (
	"net/http/httptest"
	"testing"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestMarshaler_writeResponse(t *testing.T) {
	m := &jsonapiMarshaler{}
	r := httptest.NewRequest("GET", "/?", nil)

	t.Run("organization", func(t *testing.T) {
		w := httptest.NewRecorder()
		m.writeResponse(w, r, &organization.Organization{
			Name: "acmeco",
		})

		got := w.Body.String()
		want := `{"data":{"id":"acmeco","type":"organizations","attributes":{"allow-force-delete-workspaces":false,"assessments-enforced":false,"collaborator-auth-policy":"","cost-estimation-enabled":false,"created-at":"0001-01-01T00:00:00Z","email":"","external-id":"","owners-team-saml-role-id":"","permissions":{"can-create-team":false,"can-create-workspace":true,"can-create-workspace-migration":false,"can-destroy":true,"can-traverse":false,"can-update":true,"can-update-api-token":false,"can-update-oauth":false,"can-update-sentinel":false},"saml-enabled":false,"send-passing-statuses-for-untriggered-speculative-plans":false,"session-remember":null,"session-timeout":null,"trial-expires-at":"0001-01-01T00:00:00Z","two-factor-conformant":false}}}`
		assert.Equal(t, want, got)
	})

	t.Run("organization list", func(t *testing.T) {
		w := httptest.NewRecorder()
		m.writeResponse(w, r, []*organization.Organization{
			{
				Name: "acmeco",
			},
		})

		got := w.Body.String()
		want := `{"data":[{"id":"acmeco","type":"organizations","attributes":{"allow-force-delete-workspaces":false,"assessments-enforced":false,"collaborator-auth-policy":"","cost-estimation-enabled":false,"created-at":"0001-01-01T00:00:00Z","email":"","external-id":"","owners-team-saml-role-id":"","permissions":{"can-create-team":false,"can-create-workspace":true,"can-create-workspace-migration":false,"can-destroy":true,"can-traverse":false,"can-update":true,"can-update-api-token":false,"can-update-oauth":false,"can-update-sentinel":false},"saml-enabled":false,"send-passing-statuses-for-untriggered-speculative-plans":false,"session-remember":null,"session-timeout":null,"trial-expires-at":"0001-01-01T00:00:00Z","two-factor-conformant":false}}]}`
		assert.Equal(t, want, got)
	})

	t.Run("organization page", func(t *testing.T) {
		w := httptest.NewRecorder()
		m.writeResponse(w, r, &resource.Page[*organization.Organization]{
			Items: []*organization.Organization{
				{
					Name: "acmeco",
				},
			},
			Pagination: &resource.Pagination{},
		})

		got := w.Body.String()
		want := `{"data":[{"id":"acmeco","type":"organizations","attributes":{"allow-force-delete-workspaces":false,"assessments-enforced":false,"collaborator-auth-policy":"","cost-estimation-enabled":false,"created-at":"0001-01-01T00:00:00Z","email":"","external-id":"","owners-team-saml-role-id":"","permissions":{"can-create-team":false,"can-create-workspace":true,"can-create-workspace-migration":false,"can-destroy":true,"can-traverse":false,"can-update":true,"can-update-api-token":false,"can-update-oauth":false,"can-update-sentinel":false},"saml-enabled":false,"send-passing-statuses-for-untriggered-speculative-plans":false,"session-remember":null,"session-timeout":null,"trial-expires-at":"0001-01-01T00:00:00Z","two-factor-conformant":false}}],"meta":{"pagination":{"current-page":0,"prev-page":null,"next-page":null,"total-pages":0,"total-count":0}}}`
		assert.Equal(t, want, got)
	})
}
