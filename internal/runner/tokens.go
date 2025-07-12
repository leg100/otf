package runner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
	"github.com/leg100/otf/internal/workspace"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

const (
	defaultJobTokenExpiry = 60 * time.Minute
)

type (
	// agentToken represents the authentication token for an agent.
	// NOTE: the cryptographic token itself is not retained.
	agentToken struct {
		ID          resource.TfeID `jsonapi:"primary,agent_tokens" db:"agent_token_id"`
		CreatedAt   time.Time
		AgentPoolID resource.TfeID `jsonapi:"attribute" json:"agent_pool_id" db:"agent_pool_id"`
		Description string         `jsonapi:"attribute" json:"description"`
	}

	CreateAgentTokenOptions struct {
		Description string `json:"description" schema:"description,required"`
	}
)

func (a *agentToken) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", a.ID.String()),
		slog.String("agent_pool_id", string(a.AgentPoolID.String())),
		slog.String("description", a.Description),
	}
	return slog.GroupValue(attrs...)
}

type tokenFactory struct {
	tokens     *tokens.Service
	runners    *Service
	workspaces *workspace.Service
}

// createJobToken constructs a job token
func (f *tokenFactory) createJobToken(jobID resource.TfeID) ([]byte, error) {
	expiry := internal.CurrentTimestamp(nil).Add(defaultJobTokenExpiry)
	return f.tokens.NewToken(jobID, tokens.WithExpiry(expiry))
}

// NewAgentToken constructs a token for an agent, returning both the
// representation of the token, and the cryptographic token itself.
func (f *tokenFactory) NewAgentToken(poolID resource.TfeID, opts CreateAgentTokenOptions) (*agentToken, []byte, error) {
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := agentToken{
		ID:          resource.NewTfeID(resource.AgentTokenKind),
		CreatedAt:   internal.CurrentTimestamp(nil),
		Description: opts.Description,
		AgentPoolID: poolID,
	}
	token, err := f.tokens.NewToken(at.ID)
	if err != nil {
		return nil, nil, err
	}
	return &at, token, nil
}

func (f *tokenFactory) newDynamicCredentialsToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	if f.tokens.PrivateKey == nil {
		return nil, errors.New("no private key has been configured")
	}
	job, err := f.runners.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	workspace, err := f.workspaces.Get(ctx, job.WorkspaceID)
	if err != nil {
		return nil, err
	}
	workspacePath := fmt.Sprintf("organization:%s:workspace:%s", job.Organization, workspace.Name)
	subject := fmt.Sprintf("%s:run_phase:%s", workspacePath, job.Phase)
	builder := jwt.NewBuilder().
		Subject(subject).
		Audience([]string{audience}).
		IssuedAt(time.Now()).
		Expiration(time.Now().Add(time.Hour)).
		Claim("terraform_organization_name", job.Organization).
		Claim("terraform_workspace_name", workspace.Name).
		Claim("terraform_workspace_id", workspace.ID).
		Claim("terraform_full_workspace", workspacePath).
		Claim("terraform_run_id", job.RunID).
		Claim("terraform_run_phase", job.Phase)
	token, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return jwt.Sign(token, jwt.WithKey(f.tokens.PrivateKey.Algorithm(), f.tokens.PrivateKey))
}
