package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of users and teams and various auth tokens
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

	// CreateSession persists a new session to the store.
	CreateSession(ctx context.Context, session *Session) error
	// GetSession retrieves a session using its token.
	GetSessionByToken(ctx context.Context, token string) (*Session, error)
	// ListSessions lists current sessions for a user
	ListSessions(ctx context.Context, userID string) ([]*Session, error)
	// DeleteSession deletes a session
	DeleteSession(ctx context.Context, token string) error

	CreateAgentToken(ctx context.Context, at *agentToken) error
	// GetAgentTokenByID retrieves agent token using its ID.
	GetAgentTokenByID(ctx context.Context, id string) (*agentToken, error)
	// GetAgentTokenByToken retrieves agent token using its cryptographic
	// authentication token.
	GetAgentTokenByToken(ctx context.Context, token string) (*agentToken, error)
	ListAgentTokens(ctx context.Context, organization string) ([]*agentToken, error)
	DeleteAgentToken(ctx context.Context, id string) error

	createRegistrySession(context.Context, *registrySession) error
	getRegistrySession(ctx context.Context, token string) (*registrySession, error)

	// listTeamMembers lists users that are members of the given team
	listTeamMembers(ctx context.Context, teamID string) ([]*User, error)

	tx(context.Context, func(db) error) error
}

// pgdb is a registry session database on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
	logr.Logger
}

func newDB(database otf.Database, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
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

func (db *pgdb) listTeamMembers(ctx context.Context, teamID string) ([]*User, error) {
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
		return callback(newDB(tx, db.Logger))
	})
}

// Registry sessions database

func (db *pgdb) createRegistrySession(ctx context.Context, session *registrySession) error {
	_, err := db.InsertRegistrySession(ctx, pggen.InsertRegistrySessionParams{
		Token:            sql.String(session.Token()),
		Expiry:           sql.Timestamptz(session.Expiry()),
		OrganizationName: sql.String(session.Organization()),
	})
	return sql.Error(err)
}

func (db *pgdb) getRegistrySession(ctx context.Context, token string) (*registrySession, error) {
	row, err := db.FindRegistrySession(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return registrySessionRow(row).toRegistrySession(), nil
}

// CreateAgentToken inserts an agent token, associating it with an organization
func (db *pgdb) CreateAgentToken(ctx context.Context, token *agentToken) error {
	_, err := db.InsertAgentToken(ctx, pggen.InsertAgentTokenParams{
		TokenID:          sql.String(token.ID()),
		Token:            sql.String(*token.Token()),
		Description:      sql.String(token.Description()),
		OrganizationName: sql.String(token.Organization()),
		CreatedAt:        sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *pgdb) ListAgentTokens(ctx context.Context, organizationName string) ([]*agentToken, error) {
	rows, err := db.FindAgentTokens(ctx, sql.String(organizationName))
	if err != nil {
		return nil, sql.Error(err)
	}
	var unmarshalled []*agentToken
	for _, r := range rows {
		unmarshalled = append(unmarshalled, agentTokenRow(r).toAgentToken())
	}
	return unmarshalled, nil
}

func (db *pgdb) GetAgentTokenByID(ctx context.Context, id string) (*agentToken, error) {
	r, err := db.FindAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

func (db *pgdb) GetAgentTokenByToken(ctx context.Context, token string) (*agentToken, error) {
	r, err := db.FindAgentTokenByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return agentTokenRow(r).toAgentToken(), nil
}

// DeleteAgentToken deletes an agent token.
func (db *pgdb) DeleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

// CreateSession inserts the session, associating it with the user.
func (db *pgdb) CreateSession(ctx context.Context, session *Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     sql.String(session.Token()),
		Address:   sql.String(session.Address()),
		Expiry:    sql.Timestamptz(session.Expiry()),
		UserID:    sql.String(session.UserID()),
		CreatedAt: sql.Timestamptz(session.CreatedAt()),
	})
	return err
}

func (db *pgdb) GetSessionByToken(ctx context.Context, token string) (*Session, error) {
	result, err := db.FindSessionByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return sessionRow(result).toSession(), nil
}

func (db *pgdb) ListSessions(ctx context.Context, userID string) ([]*Session, error) {
	result, err := db.FindSessionsByUserID(ctx, sql.String(userID))
	if err != nil {
		return nil, err
	}
	var sessions []*Session
	for _, row := range result {
		sessions = append(sessions, sessionRow(row).toSession())
	}
	return sessions, nil
}

// DeleteSession deletes a user's session from the DB.
func (db *pgdb) DeleteSession(ctx context.Context, token string) error {
	_, err := db.DeleteSessionByToken(ctx, sql.String(token))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) startSessionExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (db *pgdb) startExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			// TODO: log errors
			_, err := db.DeleteExpiredRegistrySessions(ctx)
			_, err := db.DeleteSessionsExpired(context.Background())
			if err != nil {
				return sql.Error(err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (db *pgdb) deleteExpired() error {
	return nil
}
