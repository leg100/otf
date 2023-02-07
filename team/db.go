package team

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a team database
type db interface {
	otf.Database

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

// pgdb is the hook database on postgres
type DB struct {
	otf.Database
}

func newDB(db otf.Database) *DB {
	return &DB{db}
}

// CreateTeam persists a team to the DB.
func (db *DB) CreateTeam(ctx context.Context, team *Team) error {
	_, err := db.InsertTeam(ctx, pggen.InsertTeamParams{
		ID:               sql.String(team.ID()),
		Name:             sql.String(team.Name()),
		CreatedAt:        sql.Timestamptz(team.CreatedAt()),
		OrganizationName: sql.String(team.Organization()),
	})
	return sql.Error(err)
}

func (pdb *DB) UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error) {
	var team *Team
	err := pdb.tx(ctx, func(tx db) error {
		var err error

		// retrieve team
		result, err := tx.FindTeamByIDForUpdate(ctx, sql.String(teamID))
		if err != nil {
			return err
		}
		team = pgRow(result).toTeam()

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
func (db *DB) GetTeam(ctx context.Context, name, organization string) (*Team, error) {
	result, err := db.FindTeamByName(ctx, sql.String(name), sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toTeam(), nil
}

// GetTeamByID retrieves a team from the DB by ID.
func (db *DB) GetTeamByID(ctx context.Context, id string) (*Team, error) {
	result, err := db.FindTeamByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(result).toTeam(), nil
}

func (db *DB) ListTeams(ctx context.Context, organization string) ([]*Team, error) {
	result, err := db.FindTeamsByOrg(ctx, sql.String(organization))
	if err != nil {
		return nil, err
	}

	var items []*Team
	for _, r := range result {
		items = append(items, pgRow(r).toTeam())
	}
	return items, nil
}

// DeleteTeam deletes a team from the DB.
func (db *DB) DeleteTeam(ctx context.Context, teamID string) error {
	_, err := db.DeleteTeamByID(ctx, sql.String(teamID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *DB) SetWorkspacePermission(ctx context.Context, workspaceID, team string, role rbac.Role) error {
	_, err := db.UpsertWorkspacePermission(ctx, pggen.UpsertWorkspacePermissionParams{
		WorkspaceID: sql.String(workspaceID),
		TeamName:    sql.String(team),
		Role:        sql.String(role.String()),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *DB) ListWorkspacePermissions(ctx context.Context, workspaceID string) ([]*otf.WorkspacePermission, error) {
	result, err := db.FindWorkspacePermissionsByID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}
	var perms []*otf.WorkspacePermission
	for _, row := range result {
		perm, err := permissionRow(row).toPermission()
		if err != nil {
			return nil, sql.Error(err)
		}
		perms = append(perms, perm)
	}
	return perms, nil
}

func (db *DB) UnsetWorkspacePermission(ctx context.Context, workspaceID, team string) error {
	_, err := db.DeleteWorkspacePermissionByID(ctx, sql.String(workspaceID), sql.String(team))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// tx constructs a new pgdb within a transaction.
func (db *DB) tx(ctx context.Context, callback func(db) error) error {
	return db.Transaction(ctx, func(tx otf.Database) error {
		return callback(newDB(tx))
	})
}

// pgRow represents the result of a database query for a team.
type pgRow struct {
	TeamID                     pgtype.Text        `json:"team_id"`
	Name                       pgtype.Text        `json:"name"`
	CreatedAt                  pgtype.Timestamptz `json:"created_at"`
	PermissionManageWorkspaces bool               `json:"permission_manage_workspaces"`
	PermissionManageVCS        bool               `json:"permission_manage_vcs"`
	PermissionManageRegistry   bool               `json:"permission_manage_registry"`
	OrganizationName           pgtype.Text        `json:"organization_name"`
}

func (row pgRow) toTeam() *Team {
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

// permissionRow represents the result of a database query for a
// workspace permission.
type permissionRow struct {
	Role         pgtype.Text          `json:"role"`
	Team         *pggen.Teams         `json:"team"`
	Organization *pggen.Organizations `json:"organization"`
}

func (row permissionRow) toPermission() (*otf.WorkspacePermission, error) {
	role, err := rbac.WorkspaceRoleFromString(row.Role.String)
	if err != nil {
		return nil, err
	}
	return &otf.WorkspacePermission{
		Role: role,
		Team: pgRow(*row.Team).toTeam(),
	}, nil
}
