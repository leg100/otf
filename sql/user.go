package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

var (
	_ otf.UserStore = (*UserDB)(nil)
)

type UserDB struct {
	*pgxpool.Pool
}

func NewUserDB(conn *pgxpool.Pool) *UserDB {
	return &UserDB{
		Pool: conn,
	}
}

// Create persists a User to the DB.
func (db UserDB) Create(ctx context.Context, user *otf.User) error {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := pggen.NewQuerier(tx)
	_, err = q.InsertUser(ctx, pggen.InsertUserParams{
		ID:        pgtype.Text{String: user.ID(), Status: pgtype.Present},
		Username:  pgtype.Text{String: user.Username(), Status: pgtype.Present},
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	})
	if err != nil {
		return err
	}
	for _, org := range user.Organizations {
		_, err = q.InsertOrganizationMembership(ctx,
			pgtype.Text{String: user.ID(), Status: pgtype.Present},
			pgtype.Text{String: org.ID(), Status: pgtype.Present},
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (db UserDB) SetCurrentOrganization(ctx context.Context, userID, orgName string) error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.UpdateUserCurrentOrganization(ctx, pggen.UpdateUserCurrentOrganizationParams{
		ID:                  pgtype.Text{String: userID, Status: pgtype.Present},
		CurrentOrganization: pgtype.Text{String: orgName, Status: pgtype.Present},
		UpdatedAt:           otf.CurrentTimestamp(),
	})
	return err
}

func (db UserDB) List(ctx context.Context) ([]*otf.User, error) {
	q := pggen.NewQuerier(db.Pool)

	result, err := q.FindUsers(ctx)
	if err != nil {
		return nil, err
	}

	var items []*otf.User
	for _, r := range result {
		user, err := otf.UnmarshalUserDBResult(otf.UserDBResult(r))
		if err != nil {
			return nil, err
		}
		items = append(items, user)
	}
	return items, nil
}

// Get retrieves a user from the DB, along with its sessions.
func (db UserDB) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	q := pggen.NewQuerier(db.Pool)

	if spec.UserID != nil {
		result, err := q.FindUserByID(ctx,
			pgtype.Text{String: *spec.UserID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.Username != nil {
		result, err := q.FindUserByUsername(ctx,
			pgtype.Text{String: *spec.Username, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationToken != nil {
		result, err := q.FindUserByAuthenticationToken(ctx,
			pgtype.Text{String: *spec.AuthenticationToken, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationTokenID != nil {
		result, err := q.FindUserByAuthenticationTokenID(ctx,
			pgtype.Text{String: *spec.AuthenticationTokenID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.SessionToken != nil {
		result, err := q.FindUserBySessionToken(ctx,
			pgtype.Text{String: *spec.SessionToken, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db UserDB) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	q := pggen.NewQuerier(db.Pool)

	_, err := q.InsertOrganizationMembership(ctx,
		pgtype.Text{String: id, Status: pgtype.Present},
		pgtype.Text{String: orgID, Status: pgtype.Present},
	)
	return err
}

func (db UserDB) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	q := pggen.NewQuerier(db.Pool)

	result, err := q.DeleteOrganizationMembership(ctx,
		pgtype.Text{String: id, Status: pgtype.Present},
		pgtype.Text{String: orgID, Status: pgtype.Present},
	)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}

// Delete deletes a user from the DB.
func (db UserDB) Delete(ctx context.Context, spec otf.UserSpec) error {
	q := pggen.NewQuerier(db.Pool)

	var result pgconn.CommandTag
	var err error
	if spec.UserID != nil {
		result, err = q.DeleteUserByID(ctx,
			pgtype.Text{String: *spec.UserID, Status: pgtype.Present},
		)
	} else if spec.Username != nil {
		result, err = q.DeleteUserByUsername(ctx,
			pgtype.Text{String: *spec.Username, Status: pgtype.Present},
		)
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return otf.ErrResourceNotFound
	}

	return nil
}
