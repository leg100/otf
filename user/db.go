package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of users and teams
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

	CreateTeam(ctx context.Context, team *Team) error
	UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error)
	GetTeam(ctx context.Context, name, organization string) (*Team, error)
	GetTeamByID(ctx context.Context, teamID string) (*Team, error)
	DeleteTeam(ctx context.Context, teamID string) error
	ListTeams(ctx context.Context, organization string) ([]*Team, error)
	// ListTeamMembers lists users that are members of the given team
	ListTeamMembers(ctx context.Context, teamID string) ([]otf.User, error)

	tx(context.Context, func(db) error) error
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
			users = append(users, userRow(r).toUser())
		}
	} else if opts.Organization != nil {
		result, err := db.FindUsersByOrganization(ctx, sql.String(*opts.Organization))
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, userRow(r).toUser())
		}
	} else {
		result, err := db.FindUsers(ctx)
		if err != nil {
			return nil, err
		}
		for _, r := range result {
			users = append(users, userRow(r).toUser())
		}
	}
	return users, nil
}

func (db *pgdb) ListTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
	result, err := db.FindUsersByTeamID(ctx, sql.String(teamID))
	if err != nil {
		return nil, err
	}

	var items []*User
	for _, r := range result {
		items = append(items, userRow(r).toUser())
	}
	return items, nil
}

// GetUser retrieves a user from the DB, along with its sessions.
func (db *pgdb) GetUser(ctx context.Context, spec UserSpec) (*User, error) {
	if spec.UserID != nil {
		result, err := db.FindUserByID(ctx, sql.String(*spec.UserID))
		if err != nil {
			return nil, err
		}
		return userRow(result).toUser(), nil
	} else if spec.Username != nil {
		result, err := db.FindUserByUsername(ctx, sql.String(*spec.Username))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else if spec.AuthenticationToken != nil {
		result, err := db.FindUserByAuthenticationToken(ctx, sql.String(*spec.AuthenticationToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
	} else if spec.SessionToken != nil {
		result, err := db.FindUserBySessionToken(ctx, sql.String(*spec.SessionToken))
		if err != nil {
			return nil, sql.Error(err)
		}
		return userRow(result).toUser(), nil
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

func (db *pgdb) AddTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.InsertTeamMembership(ctx, sql.String(userID), sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) RemoveTeamMembership(ctx context.Context, userID, teamID string) error {
	_, err := db.DeleteTeamMembership(ctx, sql.String(userID), sql.String(teamID))
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

// CreateTeam persists a team to the DB.
func (db *pgdb) CreateTeam(ctx context.Context, team *Team) error {
	_, err := db.InsertTeam(ctx, pggen.InsertTeamParams{
		ID:               sql.String(team.ID()),
		Name:             sql.String(team.Name()),
		CreatedAt:        sql.Timestamptz(team.CreatedAt()),
		OrganizationName: sql.String(team.Organization()),
	})
	return sql.Error(err)
}

func (pdb *pgdb) UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error) {
	var team *Team
	err := pdb.tx(ctx, func(tx db) error {
		var err error

		// retrieve team
		result, err := tx.FindTeamByIDForUpdate(ctx, sql.String(teamID))
		if err != nil {
			return err
		}
		team = teamRow(result).toTeam()

		// update team
		if err := fn(team); err != nil {
			return err
		}
		// persist update
		_, err = tx.UpdateTeamByID(ctx, pggen.UpdateTeamByIDParams{
			PermissionManageWorkspaces: team.OrganizationAccess().ManageWorkspaces,
			PermissionManageVCS:        team.OrganizationAccess().ManageVCS,
			PermissionManageRegistry:   team.OrganizationAccess().ManageRegistry,
			TeamID:                     sql.String(teamID),
		})
		if err != nil {
			return err
		}
		return nil
	})
	return team, err
}

// GetTeam retrieves a team from the DB by name
func (db *pgdb) GetTeam(ctx context.Context, name, organization string) (*Team, error) {
	result, err := db.FindTeamByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

// GetTeamByID retrieves a team from the DB by ID.
func (db *pgdb) GetTeamByID(ctx context.Context, id string) (*Team, error) {
	result, err := db.FindTeamByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return teamRow(result).toTeam(), nil
}

func (db *pgdb) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	result, err := db.FindTeamsByOrg(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}

	var items []*Team
	for _, r := range result {
		items = append(items, teamRow(r).toTeam())
	}
	return items, nil
}

// DeleteTeam deletes a team from the DB.
func (db *pgdb) DeleteTeam(ctx context.Context, teamID string) error {
	_, err := db.DeleteTeamByID(ctx, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(db) error) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		return callback(newPGDB(tx))
	})
}

// userRow represents the result of a database query for a user.
type userRow struct {
	UserID        pgtype.Text        `json:"user_id"`
	Username      pgtype.Text        `json:"username"`
	CreatedAt     pgtype.Timestamptz `json:"created_at"`
	UpdatedAt     pgtype.Timestamptz `json:"updated_at"`
	Organizations []string           `json:"organizations"`
	Teams         []pggen.Teams      `json:"teams"`
}

func (row userRow) toUser() *User {
	user := User{
		id:        row.UserID.String,
		createdAt: row.CreatedAt.Time.UTC(),
		updatedAt: row.UpdatedAt.Time.UTC(),
		username:  row.Username.String,
	}
	// avoid assigning empty slice to ensure equality in tests
	if len(row.Organizations) > 0 {
		user.organizations = row.Organizations
	}
	for _, tr := range row.Teams {
		user.teams = append(user.teams, teamRow(tr).toTeam())
	}

	return &user
}

// teamRow represents the result of a database query for a team.
type teamRow struct {
	TeamID                     pgtype.Text        `json:"team_id"`
	Name                       pgtype.Text        `json:"name"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	PermissionManageWorkspaces bool               `json:"permission_manage_workspaces"`
	PermissionManageVCS        bool               `json:"permission_manage_vcs"`
	PermissionManageRegistry   bool               `json:"permission_manage_registry"`
	OrganizationName           pgtype.Text        `json:"organization_name"`
}

func (row teamRow) toTeam() *Team {
	return &Team{
		id:           row.TeamID.String,
		createdAt:    row.CreatedAt.Time.UTC(),
		name:         row.Name.String,
		organization: row.OrganizationName.String,
		access: OrganizationAccess{
			ManageWorkspaces: row.PermissionManageWorkspaces,
			ManageVCS:        row.PermissionManageVCS,
			ManageRegistry:   row.PermissionManageRegistry,
		},
	}
}

