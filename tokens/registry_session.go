package tokens

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	defaultRegistrySessionExpiry = 10 * time.Minute
)

type (
	// RegistrySession provides access to the module registry for a short period.
	// Intended for use with the terraform binary, which needs authenticated access
	// to the registry in order to retrieve modules.
	RegistrySession struct {
		Organization string
	}

	CreateRegistryTokenOptions struct {
		Organization *string    // required organization
		RunID        *string    // required ID of run that is accessing the registry
		Expiry       *time.Time // optionally override expiry
	}
)

func NewRegistrySessionFromJWT(token jwt.Token) (*RegistrySession, error) {
	org, ok := token.Get("organization")
	if !ok {
		return nil, fmt.Errorf("missing claim: organization")
	}
	return &RegistrySession{Organization: org.(string)}, nil
}

func (t *RegistrySession) String() string { return "registry-session" }
func (t *RegistrySession) ID() string     { return "registry-session" }

func (t *RegistrySession) Organizations() []string { return nil }

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
