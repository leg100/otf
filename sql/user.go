package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/leg100/otf"
)

var (
	_ otf.UserStore = (*UserDB)(nil)
)

type userRow interface {
	GetUserID() *string
	GetUsername() *string
	GetCurrentOrganization() *string

	Timestamps
}

type UserDB struct {
	*pgx.Conn
}

func NewUserDB(conn *pgx.Conn, cleanupInterval time.Duration) *UserDB {
	udb := &UserDB{
		Conn: conn,
	}
	if cleanupInterval > 0 {
		go udb.startCleanup(cleanupInterval)
	}
	return udb
}

// Create persists a User to the DB.
func (db UserDB) Create(ctx context.Context, user *otf.User) error {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	_, err = q.InsertUser(ctx, InsertUserParams{
		ID:                  &user.ID,
		Username:            &user.Username,
		CurrentOrganization: user.CurrentOrganization,
	})
	if err != nil {
		return nil
	}

	for _, org := range user.Organizations {
		_, err = q.InsertOrganizationMembership(ctx, &user.ID, &org.ID)
		if err != nil {
			return nil
		}
	}

	return nil
}

// Update persists changes to the provided user object to the backend. The spec
// identifies the user to update.
func (db UserDB) Update(ctx context.Context, spec otf.UserSpec, updated *otf.User) error {
	existing, err := getUser(ctx, db.DB, spec)
	if err != nil {
		return err
	}

	// User's organization_memberships are updated separately in a many-to-many
	// table.
	if err := updateOrganizationMemberships(ctx, db.DB, existing, updated); err != nil {
		return err
	}

	updateBuilder := psql.
		Update("users").
		Where("user_id = ?", existing.ID)

	var modified bool

	if existing.CurrentOrganization != updated.CurrentOrganization {
		updateBuilder = updateBuilder.Set("current_organization", updated.CurrentOrganization)
		modified = true
	}

	if !modified {
		return nil
	}

	sql, args, err := updateBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

func (db UserDB) List(ctx context.Context) ([]*otf.User, error) {
	q := NewQuerier(db.Conn)

	result, err := q.FindUsers(ctx)
	if err != nil {
		return nil, err
	}

	var users []*otf.User
	for _, r := range result {
		users = append(users, convertUser(r))
	}
	return users, nil
}

// Get retrieves a user from the DB, along with its sessions.
func (db UserDB) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	return getUser(ctx, NewQuerier(db.Conn), spec)
}

func (db UserDB) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	q := NewQuerier(db.Conn)

	_, err := q.InsertOrganizationMembership(ctx, &id, &orgID)
	return err
}

func (db UserDB) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	q := NewQuerier(db.Conn)

	_, err := q.DeleteOrganizationMembership(ctx, &id, &orgID)
	return err
}

// Delete deletes a user from the DB.
func (db UserDB) Delete(ctx context.Context, spec otf.UserSpec) error {
	q := NewQuerier(db.Conn)

	var result pgconn.CommandTag
	var err error

	if spec.UserID != nil {
		result, err = q.DeleteUserByID(ctx, spec.UserID)
	} else if spec.Username != nil {
		result, err = q.DeleteUserByUsername(ctx, spec.Username)
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

// CreateToken inserts the token, associating it with the user.
func (db UserDB) CreateToken(ctx context.Context, token *otf.Token) error {
	sql, args, err := db.BindNamed(insertTokenSQL, token)
	if err != nil {
		return err
	}

	_, err = db.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// DeleteToken deletes a user's token from the DB.
func (db UserDB) DeleteToken(ctx context.Context, id string) error {
	_, err := db.Exec("DELETE FROM tokens WHERE token_id = $1", id)
	if err != nil {
		return fmt.Errorf("unable to delete token: %w", err)
	}

	return nil
}

func getUser(ctx context.Context, q *DBQuerier, spec otf.UserSpec) (*otf.User, error) {
	if spec.UserID != nil {
		result, err := q.FindUserByID(ctx, spec.UserID)
		if err != nil {
			return nil, err
		}
		return convertUser(result), nil
	} else if spec.Username != nil {
		result, err := q.FindUserByUsername(ctx, spec.Username)
		if err != nil {
			return nil, err
		}
		return convertUser(result), nil
	} else if spec.AuthenticationToken != nil {
		result, err := q.FindUserByAuthenticationToken(ctx, spec.AuthenticationToken)
		if err != nil {
			return nil, err
		}
		return convertUser(result), nil
	} else if spec.AuthenticationTokenID != nil {
		result, err := q.FindUserByAuthenticationTokenID(ctx, spec.AuthenticationTokenID)
		if err != nil {
			return nil, err
		}
		return convertUser(result), nil
	} else if spec.SessionToken != nil {
		result, err := q.FindUserBySessionToken(ctx, spec.SessionToken)
		if err != nil {
			return nil, err
		}
		return convertUser(result), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func convertUser(row userRow) *otf.User {
	user := otf.User{
		ID:                  *row.GetUserID(),
		Timestamps:          convertTimestamps(row),
		Username:            *row.GetUsername(),
		CurrentOrganization: row.GetCurrentOrganization(),
	}

	return &user
}
