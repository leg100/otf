package tokens

import (
	"context"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	defaultRunTokenExpiry = 10 * time.Minute
)

type (
	// RunToken is a short-lived token providing a terraform run with access to
	// resources, in particular access to the registry to retrieve modules.
	RunToken struct {
		Organization string
	}

	CreateRunTokenOptions struct {
		Organization *string    // Organization of run. Required.
		RunID        *string    // ID of run. Required.
		Expiry       *time.Time // Override expiry. Optional.
	}

	RunTokenService interface {
		CreateRunToken(ctx context.Context, opts CreateRunTokenOptions) ([]byte, error)
	}
)

func NewRunTokenFromJWT(token jwt.Token) (*RunToken, error) {
	org, ok := token.Get("organization")
	if !ok {
		return nil, fmt.Errorf("missing claim: organization")
	}
	return &RunToken{Organization: org.(string)}, nil
}

func (t *RunToken) String() string { return "run-token" }
func (t *RunToken) ID() string     { return "run-token" }

func (t *RunToken) Organizations() []string { return nil }

func (t *RunToken) IsSiteAdmin() bool   { return true }
func (t *RunToken) IsOwner(string) bool { return true }

func (t *RunToken) CanAccessSite(action rbac.Action) bool {
	return false
}

func (t *RunToken) CanAccessOrganization(action rbac.Action, name string) bool {
	// run token is only allowed read-access to its organization's module registry
	switch action {
	case rbac.GetModuleAction, rbac.ListModulesAction:
		return t.Organization == name
	default:
		return false
	}
}

func (t *RunToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	return false
}

func (a *service) CreateRunToken(ctx context.Context, opts CreateRunTokenOptions) ([]byte, error) {
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing organization")
	}
	if opts.RunID == nil {
		return nil, fmt.Errorf("missing run ID")
	}

	subject, err := a.organization.CanAccess(ctx, rbac.CreateRunTokenAction, *opts.Organization)
	if err != nil {
		return nil, err
	}

	expiry := internal.CurrentTimestamp().Add(defaultRunTokenExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := newToken(newTokenOptions{
		key:     a.key,
		subject: *opts.RunID,
		kind:    runTokenKind,
		expiry:  &expiry,
		claims: map[string]string{
			"organization": *opts.Organization,
		},
	})
	if err != nil {
		return nil, err
	}

	a.V(2).Info("created run token", "subject", subject, "run")

	return token, nil
}
