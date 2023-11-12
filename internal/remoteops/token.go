package remoteops

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/http/html"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/otf/internal/tokens"
)

const AgentTokenKind tokens.Kind = "agent_token"

type (
	// AgentToken represents the authentication token for an external agent.
	// NOTE: the cryptographic token itself is not retained.
	AgentToken struct {
		ID           string `jsonapi:"primary,agent_tokens"`
		CreatedAt    time.Time
		Description  string `jsonapi:"attribute" json:"description"`
		Organization string `jsonapi:"attribute" json:"organization_name"`
	}

	CreateAgentTokenOptions struct {
		Organization string `json:"organization_name" schema:"organization_name,required"`
		Description  string `json:"description" schema:"description,required"`
	}

	AgentTokenService interface {
		CreateAgentToken(ctx context.Context, options CreateAgentTokenOptions) ([]byte, error)
		GetAgentToken(ctx context.Context, id string) (*AgentToken, error)
		ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error)
		DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error)
	}

	// tokenFactory constructs agent tokens
	tokenFactory struct {
		tokens.TokensService
	}
)

// NewAgentToken constructs a token for an external agent, returning both the
// representation of the token, and the cryptographic token itself.
//
// TODO(@leg100): Unit test this.
func (f *tokenFactory) NewAgentToken(opts CreateAgentTokenOptions) (*AgentToken, []byte, error) {
	if opts.Organization == "" {
		return nil, nil, fmt.Errorf("organization name cannot be an empty string")
	}
	if opts.Description == "" {
		return nil, nil, fmt.Errorf("description cannot be an empty string")
	}
	at := AgentToken{
		ID:           internal.NewID("at"),
		CreatedAt:    internal.CurrentTimestamp(nil),
		Description:  opts.Description,
		Organization: opts.Organization,
	}
	token, err := f.NewToken(tokens.NewTokenOptions{
		Subject: at.ID,
		Kind:    AgentTokenKind,
		Claims: map[string]string{
			"organization": opts.Organization,
		},
	})
	if err != nil {
		return nil, nil, err
	}
	return &at, token, nil
}

func (t *AgentToken) String() string      { return t.ID }
func (t *AgentToken) IsSiteAdmin() bool   { return true }
func (t *AgentToken) IsOwner(string) bool { return true }

func (t *AgentToken) Organizations() []string { return []string{t.Organization} }

func (*AgentToken) CanAccessSite(action rbac.Action) bool {
	// agent cannot carry out site-level actions
	return false
}

func (*AgentToken) CanAccessTeam(rbac.Action, string) bool {
	// agent cannot carry out team-level actions
	return false
}

func (t *AgentToken) CanAccessOrganization(action rbac.Action, name string) bool {
	return t.Organization == name
}

func (t *AgentToken) CanAccessWorkspace(action rbac.Action, policy internal.WorkspacePolicy) bool {
	// agent can access anything within its organization
	return t.Organization == policy.Organization
}

// AgentFromContext retrieves an agent token from a context
func AgentFromContext(ctx context.Context) (*AgentToken, error) {
	subj, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return nil, err
	}
	agent, ok := subj.(*AgentToken)
	if !ok {
		return nil, fmt.Errorf("subject found in context but it is not an agent")
	}
	return agent, nil
}

type (
	tokensService struct {
		logr.Logger

		db           *pgdb
		organization internal.Authorizer
		api          *api
		web          *webHandlers

		*tokenFactory
	}

	ServiceOptions struct {
		logr.Logger
		*sql.DB
		html.Renderer
		*tfeapi.Responder
		tokens.TokensService
	}
)

func NewService(opts ServiceOptions) *tokensService {
	svc := &tokensService{
		Logger: opts.Logger,
		db:     &pgdb{DB: opts.DB},
		tokenFactory: &tokenFactory{
			TokensService: opts.TokensService,
		},
		organization: &organization.Authorizer{Logger: opts.Logger},
	}
	svc.api = &api{
		Responder: opts.Responder,
		svc:       svc,
	}
	svc.web = &webHandlers{
		Renderer: opts.Renderer,
		svc:      svc,
	}
	// Register with auth middleware the agent token kind and a means of
	// retrieving an AgentToken corresponding to token's subject.
	opts.TokensService.RegisterKind(AgentTokenKind, func(ctx context.Context, tokenID string) (internal.Subject, error) {
		return svc.GetAgentToken(ctx, tokenID)
	})
	return svc
}

func (a *tokensService) AddHandlers(r *mux.Router) {
	a.api.addHandlers(r)
	a.web.addHandlers(r)
}

func (a *tokensService) GetAgentToken(ctx context.Context, tokenID string) (*AgentToken, error) {
	at, err := a.db.getAgentTokenByID(ctx, tokenID)
	if err != nil {
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}
	a.V(9).Info("retrieved agent token", "organization", at.Organization, "id", at.ID)
	return at, nil
}

func (a *tokensService) CreateAgentToken(ctx context.Context, opts CreateAgentTokenOptions) ([]byte, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.CreateAgentTokenAction, opts.Organization)
	if err != nil {
		return nil, err
	}

	at, token, err := a.NewAgentToken(opts)
	if err != nil {
		return nil, err
	}
	if err := a.db.createAgentToken(ctx, at); err != nil {
		a.Error(err, "creating agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
		return nil, err
	}
	a.V(0).Info("created agent token", "organization", opts.Organization, "id", at.ID, "subject", subject)
	return token, nil
}

func (a *tokensService) ListAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	subject, err := a.organization.CanAccess(ctx, rbac.ListAgentTokensAction, organization)
	if err != nil {
		return nil, err
	}

	tokens, err := a.db.listAgentTokens(ctx, organization)
	if err != nil {
		a.Error(err, "listing agent tokens", "organization", organization, "subject", subject)
		return nil, err
	}
	a.V(9).Info("listed agent tokens", "organization", organization, "subject", subject)
	return tokens, nil
}

func (a *tokensService) DeleteAgentToken(ctx context.Context, id string) (*AgentToken, error) {
	// retrieve agent token first in order to get organization for authorization
	at, err := a.db.getAgentTokenByID(ctx, id)
	if err != nil {
		// we can't reveal any info because all we have is the
		// authentication token which is sensitive.
		a.Error(err, "retrieving agent token", "token", "******")
		return nil, err
	}

	subject, err := a.organization.CanAccess(ctx, rbac.DeleteAgentTokenAction, at.Organization)
	if err != nil {
		return nil, err
	}

	if err := a.db.deleteAgentToken(ctx, id); err != nil {
		a.Error(err, "deleting agent token", "agent token", at, "subject", subject)
		return nil, err
	}
	a.V(0).Info("deleted agent token", "agent token", at, "subject", subject)
	return at, nil
}

type agentTokenRow struct {
	TokenID          pgtype.Text        `json:"token_id"`
	CreatedAt        pgtype.Timestamptz `json:"created_at"`
	Description      pgtype.Text        `json:"description"`
	OrganizationName pgtype.Text        `json:"organization_name"`
}

func (row agentTokenRow) toAgentToken() *AgentToken {
	return &AgentToken{
		ID:           row.TokenID.String,
		CreatedAt:    row.CreatedAt.Time.UTC(),
		Description:  row.Description.String,
		Organization: row.OrganizationName.String,
	}
}

// pgdb stores agent tokens in a postgres database
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (db *pgdb) createAgentToken(ctx context.Context, token *AgentToken) error {
	_, err := db.Conn(ctx).InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID),
		Description:      sql.String(token.Description),
		OrganizationName: sql.String(token.Organization),
		CreatedAt:        sql.Timestamptz(token.CreatedAt.UTC()),
	})
	return err
}

func (db *pgdb) getAgentTokenByID(ctx context.Context, id string) (*AgentToken, error) {
	r, err := db.Conn(ctx).FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *pgdb) listAgentTokens(ctx context.Context, organization string) ([]*AgentToken, error) {
	rows, err := db.Conn(ctx).FindAgentTokens(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	tokens := make([]*AgentToken, len(rows))
	for i, r := range rows {
		tokens[i] = agentTokenRow(r).toAgentToken()
	}
	return tokens, nil
}

func (db *pgdb) deleteAgentToken(ctx context.Context, id string) error {
	_, err := db.Conn(ctx).DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}
