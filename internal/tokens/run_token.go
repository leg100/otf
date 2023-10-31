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
	// RunToken is a short-lived token providing a remote operation which
	// permissions to perform the operation, e.g. retrieving and updating state,
	// pulling modules, etc.
	RunToken struct {
		Organization string
	}

	CreateRunTokenOptions struct {
		Organization *string    `json:"organization"` // Organization of run. Required.
		RunID        *string    `json:"run_id"`       // ID of run. Required.
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
	switch action {
	case rbac.GetOrganizationAction, rbac.GetEntitlementsAction, rbac.GetModuleAction, rbac.ListModulesAction:
		return t.Organization == name
	default:
		return false
	}
}

func (t *RunToken) CanAccessTeam(rbac.Action, string) bool {
	// Can't access team level actions
	return false
}

func (t *RunToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// run token is allowed the retrieve the state of the workspace only if:
	// (a) workspace is in the same organization as run token
	// (b) workspace has enabled global remote state (permitting organization-wide
	// state sharing).
	switch action {
	case rbac.GetWorkspaceAction, rbac.GetStateVersionAction, rbac.DownloadStateAction:
		if t.Organization == policy.Organization && policy.GlobalRemoteState {
			return true
		}
	}
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

	expiry := internal.CurrentTimestamp(nil).Add(defaultRunTokenExpiry)
	if opts.Expiry != nil {
		expiry = *opts.Expiry
	}

	token, err := NewToken(NewTokenOptions{
		key:     a.key,
		Subject: *opts.RunID,
		Kind:    runTokenKind,
		Expiry:  &expiry,
		Claims: map[string]string{
			"organization": *opts.Organization,
		},
	})
	if err != nil {
		return nil, err
	}

	a.V(2).Info("created run token", "subject", subject, "run")

	return token, nil
}
