package team

import (
	"context"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/tokens"
)

const TeamTokenKind tokens.Kind = "team_token"

type (
	// Token provides information about an API token for a team.
	Token struct {
		ID        string
		CreatedAt time.Time

		// Token belongs to a team
		TeamID string
		// Optional expiry.
		Expiry *time.Time
	}

	// CreateTokenOptions are options for creating an team token via the service
	// endpoint
	CreateTokenOptions struct {
		TeamID string
		Expiry *time.Time
	}

	teamTokenService interface {
		// CreateTeamToken creates a team token.
		CreateTeamToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error)
		// GetTeamToken gets the team token. If a token does not
		// exist, then nil is returned without an error.
		GetTeamToken(ctx context.Context, teamID string) (*Token, error)
		// DeleteTeamToken deletes a team token.
		DeleteTeamToken(ctx context.Context, tokenID string) error
	}

	teamTokenFactory struct {
		tokens *tokens.Service
	}
)

func (f *teamTokenFactory) NewTeamToken(opts CreateTokenOptions) (*Token, []byte, error) {
	tt := Token{
		ID:        internal.NewID("tt"),
		CreatedAt: internal.CurrentTimestamp(nil),
		TeamID:    opts.TeamID,
		Expiry:    opts.Expiry,
	}
	token, err := f.tokens.NewToken(tokens.NewTokenOptions{
		Subject: tt.ID,
		Kind:    TeamTokenKind,
		Expiry:  opts.Expiry,
	})
	if err != nil {
		return nil, nil, err
	}
	return &tt, token, nil
}

func (t *Token) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", t.ID),
		slog.String("team_id", t.TeamID),
	}
	if t.Expiry != nil {
		attrs = append(attrs, slog.Time("expiry", *t.Expiry))
	}
	return slog.GroupValue(attrs...)
}

func (a *service) CreateTeamToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error) {
	_, err := a.team.CanAccess(ctx, rbac.CreateTeamTokenAction, opts.TeamID)
	if err != nil {
		return nil, nil, err
	}

	tt, token, err := a.NewTeamToken(opts)
	if err != nil {
		a.Error(err, "constructing team token", "team_id", opts.TeamID)
		return nil, nil, err
	}

	if err := a.db.createTeamToken(ctx, tt); err != nil {
		a.Error(err, "creating team token", "token", tt)
		return nil, nil, err
	}

	a.V(0).Info("created team token", "token", tt)

	return tt, token, nil
}

func (a *service) GetTeamToken(ctx context.Context, teamID string) (*Token, error) {
	_, err := a.team.CanAccess(ctx, rbac.GetTeamTokenAction, teamID)
	if err != nil {
		return nil, err
	}
	return a.db.getTeamTokenByTeamID(ctx, teamID)
}

func (a *service) DeleteTeamToken(ctx context.Context, teamID string) error {
	_, err := a.team.CanAccess(ctx, rbac.DeleteTeamTokenAction, teamID)
	if err != nil {
		return err
	}

	if err := a.db.deleteTeamToken(ctx, teamID); err != nil {
		a.Error(err, "deleting team token", "team", teamID)
		return err
	}

	a.V(0).Info("deleted team token", "team", teamID)

	return nil
}
