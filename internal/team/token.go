package team

import (
	"context"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/tokens"
)

const TeamTokenKind resource.Kind = "tt"

type (
	// Token provides information about an API token for a team.
	Token struct {
		resource.ID

		CreatedAt time.Time
		// Token belongs to a team
		TeamID resource.ID
		// Optional expiry.
		Expiry *time.Time
	}

	// CreateTokenOptions are options for creating an team token via the service
	// endpoint
	CreateTokenOptions struct {
		TeamID resource.ID
		Expiry *time.Time
	}

	teamTokenFactory struct {
		tokens *tokens.Service
	}
)

func (f *teamTokenFactory) NewTeamToken(opts CreateTokenOptions) (*Token, []byte, error) {
	tt := Token{
		ID:        resource.NewID(TeamTokenKind),
		CreatedAt: internal.CurrentTimestamp(nil),
		TeamID:    opts.TeamID,
		Expiry:    opts.Expiry,
	}
	var newTokenOptions []tokens.NewTokenOption
	if opts.Expiry != nil {
		newTokenOptions = append(newTokenOptions, tokens.WithExpiry(*opts.Expiry))
	}
	token, err := f.tokens.NewToken(tt.ID, newTokenOptions...)
	if err != nil {
		return nil, nil, err
	}
	return &tt, token, nil
}

func (t *Token) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", t.ID.String()),
		slog.String("team_id", t.TeamID.String()),
	}
	if t.Expiry != nil {
		attrs = append(attrs, slog.Time("expiry", *t.Expiry))
	}
	return slog.GroupValue(attrs...)
}

func (a *Service) CreateTeamToken(ctx context.Context, opts CreateTokenOptions) (*Token, []byte, error) {
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

func (a *Service) GetTeamToken(ctx context.Context, teamID resource.ID) (*Token, error) {
	_, err := a.team.CanAccess(ctx, rbac.GetTeamTokenAction, teamID)
	if err != nil {
		return nil, err
	}
	return a.db.getTeamTokenByTeamID(ctx, teamID)
}

func (a *Service) DeleteTeamToken(ctx context.Context, teamID resource.ID) error {
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
