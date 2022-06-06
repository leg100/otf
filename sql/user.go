package sql

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

// CreateUser persists a User to the DB.
func (db *DB) CreateUser(ctx context.Context, user *otf.User) error {
	tx, err := db.Begin(ctx)
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

func (db *DB) ListUsers(ctx context.Context) ([]*otf.User, error) {
	result, err := db.FindUsers(ctx)
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

// GetUser retrieves a user from the DB, along with its sessions.
func (db *DB) GetUser(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	if spec.UserID != nil {
		result, err := db.FindUserByID(ctx,
			pgtype.Text{String: *spec.UserID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx,
			pgtype.Text{String: *spec.Username, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx,
			pgtype.Text{String: *spec.AuthenticationToken, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationTokenID != nil {
		result, err := db.FindUserByAuthenticationTokenID(ctx,
			pgtype.Text{String: *spec.AuthenticationTokenID, Status: pgtype.Present},
		)
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx,
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

func (db *DB) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.InsertOrganizationMembership(ctx,
		pgtype.Text{String: id, Status: pgtype.Present},
		pgtype.Text{String: orgID, Status: pgtype.Present},
	)
	return err
}

func (db *DB) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	result, err := db.DeleteOrganizationMembership(ctx,
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

// DeleteUser deletes a user from the DB.
func (db *DB) DeleteUser(ctx context.Context, spec otf.UserSpec) error {
	var result pgconn.CommandTag
	var err error
	if spec.UserID != nil {
		result, err = db.DeleteUserByID(ctx,
			pgtype.Text{String: *spec.UserID, Status: pgtype.Present},
		)
	} else if spec.Username != nil {
		result, err = db.DeleteUserByUsername(ctx,
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
