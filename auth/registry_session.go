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

type (
	// RegistrySession provides access to the module registry for a short period.
	// Intended for use with the terraform binary, which needs authenticated access
	// to the registry in order to retrieve modules.
	RegistrySession struct {
		Token        string
		Expiry       time.Time
		Organization string
	}

	CreateRegistrySessionOptions struct {
		Organization *string    // required organization
		Expiry       *time.Time // optionally override expiry
	}
)

func NewRegistrySession(opts CreateRegistrySessionOptions) (*RegistrySession, error) {
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing organization")
	}
	token, err := otf.GenerateAuthToken("registry")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	expiry := otf.CurrentTimestamp().Add(defaultRegistrySessionExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry // override expiry
	}
	return &RegistrySession{
		Token:        token,
		Expiry:       expiry,
		Organization: *opts.Organization,
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

func (t *RegistrySession) MarshalLog() any {
	return struct {
		Token, Organization string
	}{
		Token:        "*****",
		Organization: t.Organization,
	}
}
