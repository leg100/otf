package registry

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
)

const (
	defaultExpiry = 10 * time.Minute
)

// Session provides access to the module registry for a short period.
// Intended for use with the terraform binary, which needs authenticated access
// to the registry in order to retrieve modules.
type Session struct {
	token        string
	expiry       time.Time
	organization string
}

func newSession(organization string) (*Session, error) {
	token, err := otf.GenerateAuthToken("registry")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	return &Session{
		token:        token,
		expiry:       otf.CurrentTimestamp().Add(defaultExpiry),
		organization: organization,
	}, nil
}

func (t *Session) String() string       { return "registry-session" }
func (t *Session) ID() string           { return "registry-session" }
func (t *Session) Token() string        { return t.token }
func (t *Session) Organization() string { return t.organization }
func (t *Session) Expiry() time.Time    { return t.expiry }

// ToJSONAPI assembles a JSON-API DTO.
func (t *Session) ToJSONAPI() any {
	return &jsonapiSession{
		Token:            t.Token(),
		OrganizationName: t.Organization(),
	}
}

func (*Session) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *Session) CanAccessOrganization(action rbac.Action, name string) bool {
	// registry session is only allowed read-access to its organization's module registry
	switch action {
	case rbac.GetModuleAction, rbac.ListModulesAction:
		return t.organization == name
	default:
		return false
	}
}

func (t *Session) CanAccessWorkspace(action rbac.Action, policy *otf.WorkspacePolicy) bool {
	return false
}

func (t *Session) MarshalLog() any {
	return struct {
		Token, Organization string
	}{
		Token:        "*****",
		Organization: t.organization,
	}
}
