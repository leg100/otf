package runner

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/workspace"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

type dynamicCredentialsTokenGenerator struct {
	privateKey jwk.Key

	tokenGeneratorJobGetter
	tokenGeneratorWorkspaceGetter
}

type tokenGeneratorJobGetter interface {
	getJob(ctx context.Context, jobID resource.TfeID) (*Job, error)
}

type tokenGeneratorWorkspaceGetter interface {
	Get(ctx context.Context, workspaceID resource.TfeID) (*workspace.Workspace, error)
}

func (g *dynamicCredentialsTokenGenerator) generateToken(ctx context.Context, jobID resource.TfeID, audience string) ([]byte, error) {
	if g.privateKey == nil {
		return nil, errors.New("no private key has been configured")
	}
	job, err := g.getJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	workspace, err := g.Get(ctx, job.WorkspaceID)
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
	return jwt.Sign(token, jwt.WithKey(g.privateKey.Algorithm(), g.privateKey))
}
