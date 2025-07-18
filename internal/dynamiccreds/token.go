package dynamiccreds

import (
	"fmt"
	"time"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/run"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

func GenerateToken(
	privateKey jwk.Key,
	issuer string,
	organization organization.Name,
	workspaceID resource.TfeID,
	workspaceName string,
	runID resource.TfeID,
	jobID resource.TfeID,
	phase run.PhaseType,
	audience string,
) ([]byte, error) {
	now := time.Now()
	workspacePath := fmt.Sprintf("organization:%s:workspace:%s", organization, workspaceName)
	subject := fmt.Sprintf("%s:run_phase:%s", workspacePath, phase)
	builder := jwt.NewBuilder().
		Subject(subject).
		Audience([]string{audience}).
		IssuedAt(now).
		Issuer(issuer).
		NotBefore(now).
		Expiration(now.Add(time.Hour)).
		Claim("terraform_organization_name", organization).
		Claim("terraform_workspace_name", workspaceName).
		Claim("terraform_workspace_id", workspaceID).
		Claim("terraform_full_workspace", workspacePath).
		Claim("terraform_run_id", runID).
		Claim("terraform_run_phase", phase)
	token, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("building token: %w", err)
	}
	signedToken, err := jwt.Sign(token, jwt.WithKey(jwa.RS256, privateKey))
	if err != nil {
		return nil, fmt.Errorf("signing token: %w", err)
	}
	return signedToken, nil
}
