package tokens

import (
	"context"
	"time"

	"log/slog"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type (
	// TeamToken provides information about an API token for a user.
	TeamToken struct {
		ID        string
		CreatedAt time.Time

		// Token belongs to an team
		TeamID string
		// Optional expiry.
		Expiry *time.Time
	}

	// CreateTeamTokenOptions are options for creating an team token via the service
	// endpoint
	CreateTeamTokenOptions struct {
		TeamID string `schema:"team_id,required"`
		Expiry *time.Time
	}

	// NewTeamTokenOptions are options for constructing a user token via the
	// constructor.
	NewTeamTokenOptions struct {
		CreateTeamTokenOptions
		Team string
		key  jwk.Key
	}

	teamTokenService interface {
		// CreateTeamToken creates a user token.
		CreateTeamToken(ctx context.Context, opts CreateTeamTokenOptions) (*TeamToken, []byte, error)
		// GetTeamToken gets the team token. If a token does not
		// exist, then nil is returned without an error.
		GetTeamToken(ctx context.Context, team string) (*TeamToken, error)
		// DeleteTeamToken deletes an team token.
		DeleteTeamToken(ctx context.Context, tokenID string) error
	}
)

func NewTeamToken(opts NewTeamTokenOptions) (*TeamToken, []byte, error) {
	tt := TeamToken{
		ID:        internal.NewID("tt"),
		CreatedAt: internal.CurrentTimestamp(),
		TeamID:    opts.Team,
		Expiry:    opts.Expiry,
	}
	token, err := NewToken(NewTokenOptions{
		key:     opts.key,
		Subject: tt.ID,
		Kind:    teamTokenKind,
		Expiry:  opts.Expiry,
	})
	if err != nil {
		return nil, nil, err
	}
	return &tt, token, nil
}

func (t *TeamToken) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", t.ID),
		slog.String("team_id", t.TeamID),
	}
	if t.Expiry != nil {
		attrs = append(attrs, slog.Time("expiry", *t.Expiry))
	}
	return slog.GroupValue(attrs...)
}

func (a *service) CreateTeamToken(ctx context.Context, opts CreateTeamTokenOptions) (*TeamToken, []byte, error) {
	_, err := a.team.CanAccess(ctx, rbac.CreateTeamTokenAction, opts.TeamID)
	if err != nil {
		return nil, nil, err
	}

	tt, token, err := NewTeamToken(NewTeamTokenOptions{
		CreateTeamTokenOptions: opts,
		Team:                   opts.TeamID,
		key:                    a.key,
	})
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

func (a *service) GetTeamToken(ctx context.Context, teamID string) (*TeamToken, error) {
	return a.db.getTeamTokenByTeamID(ctx, teamID)
}

func (a *service) DeleteTeamToken(ctx context.Context, teamID string) error {
	_, err := a.team.CanAccess(ctx, rbac.CreateTeamTokenAction, teamID)
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
