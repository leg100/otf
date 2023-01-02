package otf

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgtype"
)

const (
	DefaultRegistrySessionExpiry = 10 * time.Minute
)

// RegistrySession provides access to the module registry for a short period.
// Intended for use with the terraform binary, which needs authenticated access
// to the registry in order to retrieve modules.
type RegistrySession struct {
	token        string
	expiry       time.Time
	organization string
}

func NewRegistrySession(organization string) (*RegistrySession, error) {
	token, err := GenerateAuthToken("registry")
	if err != nil {
		return nil, fmt.Errorf("generating token: %w", err)
	}
	return &RegistrySession{
		token:        token,
		expiry:       CurrentTimestamp().Add(DefaultRegistrySessionExpiry),
		organization: organization,
	}, nil
}

func (t *RegistrySession) String() string       { return "registry-session" }
func (t *RegistrySession) ID() string           { return "registry-session" }
func (t *RegistrySession) Token() string        { return t.token }
func (t *RegistrySession) Organization() string { return t.organization }
func (t *RegistrySession) Expiry() time.Time    { return t.expiry }

func (*RegistrySession) CanAccessSite(action Action) bool {
	// terraform cannot carry out site-level actions
	return false
}

func (t *RegistrySession) CanAccessOrganization(action Action, name string) bool {
	// terraform is allowed to carry out actions on the given organization.
	return t.organization == name
}

func (t *RegistrySession) CanAccessWorkspace(action Action, policy *WorkspacePolicy) bool {
	// terraform can access anything within the organization
	return t.organization == policy.OrganizationName
}

func (t *RegistrySession) MarshalLog() any {
	return struct {
		Token, Organization string
	}{
		Token:        "*****",
		Organization: t.organization,
	}
}

type RegistrySessionService interface {
	CreateRegistrySession(ctx context.Context, organization string) (*RegistrySession, error)
	GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error)
}

type RegistrySessionStore interface {
	CreateRegistrySession(ctx context.Context, session *RegistrySession) error
	GetRegistrySession(ctx context.Context, token string) (*RegistrySession, error)
}

type RegistrySessionRow struct {
	Token            pgtype.Text        `json:"token"`
	Expiry           pgtype.Timestamptz `json:"expiry"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

func UnmarshalRegistrySessionRow(result RegistrySessionRow) *RegistrySession {
	return &RegistrySession{
		token:        result.Token.String,
		expiry:       result.Expiry.Time.UTC(),
		organization: result.OrganizationName.String,
	}
}
