package auth

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
	"github.com/leg100/otf/rbac"
)

const (
	defaultRegistrySessionExpiry = 10 * time.Minute
)

// Session provides access to the module registry for a short period.
// Intended for use with the terraform binary, which needs authenticated access
// to the registry in order to retrieve modules.
type RegistrySession struct {
	Token        string
	Expiry       time.Time
	Organization string
}

func NewRegistrySession(organization string) (*RegistrySession, error) {
	token, err := otf.GenerateAuthToken("registry")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	return &RegistrySession{
		Token:        token,
		Expiry:       otf.CurrentTimestamp().Add(defaultRegistrySessionExpiry),
		Organization: organization,
	}, nil
}

func (t *RegistrySession) String() string { return "registry-session" }
func (t *RegistrySession) ID() string     { return "registry-session" }

// ToJSONAPI assembles a JSON-API DTO.
func (t *RegistrySession) ToJSONAPI() any {
	return &jsonapi.RegistrySession{
		Token:            t.Token,
		OrganizationName: t.Organization,
	}
}

func (t *RegistrySession) ListOrganizations() []string { return nil }

func (t *RegistrySession) IsSiteAdmin() bool   { return true }
func (t *RegistrySession) IsOwner(string) bool { return true }

func (t *RegistrySession) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *RegistrySession) CanAccessOrganization(action rbac.Action, name string) bool {
	// registry session is only allowed read-access to its organization's module registry
	switch action {
	case rbac.GetModuleAction, rbac.ListModulesAction:
		return t.Organization == name
	default:
		return false
	}
}

func (t *RegistrySession) CanAccessWorkspace(action rbac.Action, policy otf.WorkspacePolicy) bool {
	return false
}

func (t *RegistrySession) MarshalLog() any {
	return struct {
		Token, Organization string
	}{
		Token:        "*****",
		Organization: t.Organization,
	}
}
