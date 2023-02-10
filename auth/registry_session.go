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
type registrySession struct {
	token        string
	expiry       time.Time
	organization string
}

func newRegistrySession(organization string) (*registrySession, error) {
	token, err := otf.GenerateAuthToken("registry")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	return &registrySession{
		token:        token,
		expiry:       otf.CurrentTimestamp().Add(defaultRegistrySessionExpiry),
		organization: organization,
	}, nil
}

func (t *registrySession) String() string       { return "registry-session" }
func (t *registrySession) ID() string           { return "registry-session" }
func (t *registrySession) Token() string        { return t.token }
func (t *registrySession) Organization() string { return t.organization }
func (t *registrySession) Expiry() time.Time    { return t.expiry }

// ToJSONAPI assembles a JSON-API DTO.
func (t *registrySession) ToJSONAPI() any {
	return &jsonapi.RegistrySession{
		Token:            t.Token(),
		OrganizationName: t.Organization(),
	}
}

func (*registrySession) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *registrySession) CanAccessOrganization(action rbac.Action, name string) bool {
	// registry session is only allowed read-access to its organization's module registry
	switch action {
	case rbac.GetModuleAction, rbac.ListModulesAction:
		return t.organization == name
	default:
		return false
	}
}

func (t *registrySession) CanAccessWorkspace(action rbac.Action, policy *otf.WorkspacePolicy) bool {
	return false
}

func (t *registrySession) MarshalLog() any {
	return struct {
		Token, Organization string
	}{
		Token:        "*****",
		Organization: t.organization,
	}
}
