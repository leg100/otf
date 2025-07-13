package dynamiccreds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type tokenGenerator struct {
	PrivateKey jwk.Key
}

type jobGetter interface {
	getJob(ctx context.Context, jobID resource.TfeID) (*Job, error)
}

func (g *tokenGenerator) generateToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
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
