package auth

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of users and teams and various auth tokens
type db interface {
	otf.DB

	CreateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, spec otf.UserSpec) error

	CreateAgentToken(ctx context.Context, at *agentToken) error
	// GetAgentTokenByID retrieves agent token using its ID.
	GetAgentTokenByID(ctx context.Context, id string) (*agentToken, error)
	// GetAgentTokenByToken retrieves agent token using its cryptographic
	// authentication token.
	GetAgentTokenByToken(ctx context.Context, token string) (*agentToken, error)

	listAgentTokens(ctx context.Context, organization string) ([]*agentToken, error)
	deleteAgentToken(ctx context.Context, id string) error

	createTeam(ctx context.Context, team *Team) error
	UpdateTeam(ctx context.Context, teamID string, fn func(*Team) error) (*Team, error)
	getTeam(ctx context.Context, name, organization string) (*Team, error)
	getTeamByID(ctx context.Context, teamID string) (*Team, error)
	deleteTeam(ctx context.Context, teamID string) error
	listTeams(ctx context.Context, organization string) ([]*Team, error)

	createRegistrySession(context.Context, *registrySession) error
	getRegistrySession(ctx context.Context, token string) (*registrySession, error)

	// createSession persists a new session to the store.
	createSession(ctx context.Context, session *Session) error
	// getSession retrieves a session using its token.
	getSessionByToken(ctx context.Context, token string) (*Session, error)
	// listSessions lists current sessions for a user
	listSessions(ctx context.Context, userID string) ([]*Session, error)
	// deleteSession deletes a session
	deleteSession(ctx context.Context, token string) error

	listTeamMembers(ctx context.Context, teamID string) ([]*User, error)

	listUsers(ctx context.Context, organization string) ([]*User, error)
	getUser(ctx context.Context, spec otf.UserSpec) (*User, error)

	addOrganizationMembership(ctx context.Context, userID, organization string) error
	removeOrganizationMembership(ctx context.Context, userID, organization string) error

	addTeamMembership(ctx context.Context, userID, teamID string) error
	removeTeamMembership(ctx context.Context, userID, teamID string) error

	tx(context.Context, func(db) error) error
}

// pgdb is a registry session database on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
	logr.Logger
}

func newDB(database otf.DB, logger logr.Logger) *pgdb {
	return &pgdb{database, logger}
}

// tx constructs a new pgdb within a transaction.
func (db *pgdb) tx(ctx context.Context, callback func(db) error) error {
	return db.Tx(ctx, func(tx otf.DB) error {
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
		Token:            sql.String(token.Token()),
		Description:      sql.String(token.Description()),
		OrganizationName: sql.String(token.Organization()),
		CreatedAt:        sql.Timestamptz(token.CreatedAt()),
	})
	return err
}

func (db *pgdb) listAgentTokens(ctx context.Context, organization string) ([]*agentToken, error) {
	rows, err := db.FindAgentTokens(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}
	var unmarshalled []*agentToken
	for _, r := range rows {
		unmarshalled = append(unmarshalled, agentTokenRow(r).toAgentToken())
	}
	return unmarshalled, nil
}

// deleteAgentToken deletes an agent token.
func (db *pgdb) deleteAgentToken(ctx context.Context, id string) error {
	_, err := db.DeleteAgentTokenByID(ctx, sql.String(id))
	if err != nil {
		return sql.Error(err)
	}
	return nil
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

func (db *pgdb) createSession(ctx context.Context, session *Session) error {
	_, err := db.InsertSession(ctx, pggen.InsertSessionParams{
		Token:     sql.String(session.Token()),
		Address:   sql.String(session.Address()),
		Expiry:    sql.Timestamptz(session.Expiry()),
		UserID:    sql.String(session.UserID()),
		CreatedAt: sql.Timestamptz(session.CreatedAt()),
	})
	return err
}

func (db *pgdb) getSessionByToken(ctx context.Context, token string) (*Session, error) {
	result, err := db.FindSessionByToken(ctx, sql.String(token))
	if err != nil {
		return nil, sql.Error(err)
	}
	return sessionRow(result).toSession(), nil
}

func (db *pgdb) listSessions(ctx context.Context, userID string) ([]*Session, error) {
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

func (db *pgdb) deleteSession(ctx context.Context, token string) error {
	_, err := db.DeleteSessionByToken(ctx, sql.String(token))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (db *pgdb) startExpirer(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := db.deleteExpired(ctx); err != nil {
				db.Error(err, "purging expired user sessions")
			}
		case <-ctx.Done():
			return
		}
	}
}

func (db *pgdb) deleteExpired(ctx context.Context) error {
	_, err := db.DeleteSessionsExpired(ctx)
	if err != nil {
		return err
	}
	_, err = db.DeleteExpiredRegistrySessions(ctx)
	if err != nil {
		return err
	}
	return nil
}
