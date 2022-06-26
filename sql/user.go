package sql

import (
	"context"
	"fmt"

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
		ID:        String(user.ID()),
		Username:  String(user.Username()),
		CreatedAt: Timestamptz(user.CreatedAt()),
		UpdatedAt: Timestamptz(user.UpdatedAt()),
	})
	if err != nil {
		return err
	}
	for _, org := range user.Organizations {
		_, err = q.InsertOrganizationMembership(ctx, String(user.ID()), String(org.ID()))
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
		result, err := db.FindUserByID(ctx, String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx, String(*spec.Username))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx, String(*spec.AuthenticationToken))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.AuthenticationTokenID != nil {
		result, err := db.FindUserByAuthenticationTokenID(ctx, String(*spec.AuthenticationTokenID))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx, String(*spec.SessionToken))
		if err != nil {
			return nil, databaseError(err)
		}
		return otf.UnmarshalUserDBResult(otf.UserDBResult(result))
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *DB) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.InsertOrganizationMembership(ctx, String(id), String(orgID))
	return err
}

func (db *DB) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.DeleteOrganizationMembership(ctx, String(id), String(orgID))
	if err != nil {
		return databaseError(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *DB) DeleteUser(ctx context.Context, spec otf.UserSpec) error {
	if spec.UserID != nil {
		_, err := db.DeleteUserByID(ctx, String(*spec.UserID))
		if err != nil {
			return databaseError(err)
		}
	} else if spec.Username != nil {
		_, err := db.DeleteUserByUsername(ctx, String(*spec.Username))
		if err != nil {
			return databaseError(err)
		}
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
	return nil
}
