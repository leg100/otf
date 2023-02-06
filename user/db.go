package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of users
type db interface {
	otf.Database

	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, spec UserSpec) (*User, error)
	// ListUsers lists users.
	ListUsers(ctx context.Context, opts UserListOptions) ([]*User, error)
	DeleteUser(ctx context.Context, spec UserSpec) error
	// AddOrganizationMembership adds a user as a member of an organization
	AddOrganizationMembership(ctx context.Context, id, orgID string) error
	// RemoveOrganizationMembership removes a user as a member of an
	// organization
	RemoveOrganizationMembership(ctx context.Context, id, orgID string) error
	// AddTeamMembership adds a user as a member of a team
	AddTeamMembership(ctx context.Context, id, teamID string) error
	// RemoveTeamMembership removes a user as a member of an
	// team
	RemoveTeamMembership(ctx context.Context, id, teamID string) error
}

// pgdb is a database of users on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

// CreateUser persists a User to the DB.
func (db *pgdb) CreateUser(ctx context.Context, user *User) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		_, err := tx.InsertUser(ctx, pggen.InsertUserParams{
			ID:        sql.String(user.ID()),
			Username:  sql.String(user.Username()),
			CreatedAt: sql.Timestamptz(user.CreatedAt()),
			UpdatedAt: sql.Timestamptz(user.UpdatedAt()),
		})
		if err != nil {
			return err
		}
		for _, org := range user.Organizations() {
			_, err = tx.InsertOrganizationMembership(ctx, sql.String(user.ID()), sql.String(org))
			if err != nil {
				return err
			}
		}
		for _, team := range user.Teams() {
			_, err = tx.InsertTeamMembership(ctx, sql.String(user.ID()), sql.String(team.ID()))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (db *pgdb) ListUsers(ctx context.Context, opts UserListOptions) ([]*User, error) {
	var users []*User
	if opts.Organization != nil && opts.TeamName != nil {
		result, err := db.FindUsersByTeam(ctx, sql.String(*opts.Organization), sql.String(*opts.TeamName))
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, pgRow(r).toUser())
		}
	} else if opts.Organization != nil {
		result, err := db.FindUsersByOrganization(ctx, sql.String(*opts.Organization))
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, pgRow(r).toUser())
		}
	} else {
		result, err := db.FindUsers(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, pgRow(r).toUser())
		}
	}
	return users, nil
}

// GetUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.UserID != nil {
		result, err := db.FindUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return pgRow(result).toUser(), nil
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toUser(), nil
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx, sql.String(*spec.AuthenticationToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toUser(), nil
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx, sql.String(*spec.SessionToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return pgRow(result).toUser(), nil
	} else {
		return nil, fmt.Errorf("unsupported user spec for retrieving user")
	}
}

func (db *pgdb) AddOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.InsertOrganizationMembership(ctx, sql.String(id), sql.String(orgID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) RemoveOrganizationMembership(ctx context.Context, id, orgID string) error {
	_, err := db.DeleteOrganizationMembership(ctx, sql.String(id), sql.String(orgID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// DeleteUser deletes a user from the DB.
func (db *pgdb) DeleteUser(ctx context.Context, spec UserSpec) error {
	if spec.UserID != nil {
		_, err := db.DeleteUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return sql.Error(err)
		}
	} else if spec.Username != nil {
		_, err := db.DeleteUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return sql.Error(err)
		}
	} else {
		return fmt.Errorf("unsupported user spec for deletion")
	}
	return nil
}

// pgRow represents the result of a database query for a user.
type pgRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []pggen.Teams      `json:"teams"`
}

func (row pgRow) toUser() *User {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	// avoid assigning empty slice to ensure equality in ./sql tests
	if len(row.Organizations) > 0 {
		user.organizations = row.Organizations
	}
	for _, tr := range row.Teams {
		user.teams = append(user.teams, otf.UnmarshalTeamResult(otf.TeamResult(tr)))
	}

	return &user
}
