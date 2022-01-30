package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.UserStore = (*UserDB)(nil)

	DefaultSessionCleanupInterval = 5 * time.Minute

	userColumns = []string{
		"user_id",
		"created_at",
		"updated_at",
		"username",
	}

	sessionColumns = []string{
		"token",
		"created_at",
		"updated_at",
		"flash",
		"address",
		"organization",
		"expiry",
		"user_id",
	}

	tokenColumns = []string{
		"token_id",
		"created_at",
		"updated_at",
		"description",
		"user_id",
	}

	insertUserSQL = `INSERT INTO users (user_id, created_at, updated_at, username)
VALUES (:user_id, :created_at, :updated_at, :username)`

	insertSessionSQL = `INSERT INTO sessions (token, flash, address, organization, created_at, updated_at, expiry, user_id)
VALUES (:token, :flash, :address, :organization, :created_at, :updated_at, :expiry, :user_id)`

	insertTokenSQL = `INSERT INTO tokens (token_id, token, created_at, updated_at, description, user_id)
VALUES (:token_id, :token, :created_at, :updated_at, :description, :user_id)`
)

type UserDB struct {
	*sqlx.DB
}

func NewUserDB(db *sqlx.DB, cleanupInterval time.Duration) *UserDB {
	udb := &UserDB{
		DB: db,
	}
	if cleanupInterval > 0 {
		go udb.startCleanup(cleanupInterval)
	}
	return udb
}

// Create persists a User to the DB.
func (db UserDB) Create(ctx context.Context, user *otf.User) error {
	sql, args, err := db.BindNamed(insertUserSQL, user)
	if err != nil {
		return err
	}
	_, err = db.Exec(sql, args...)
	if err != nil {
		return err
	}

	if err := addOrganizationMemberships(ctx, db.DB, user, user.Organizations); err != nil {
		return err
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

	if err := updateOrganizationMemberships(ctx, db.DB, existing, updated); err != nil {
		return err
	}

	return nil
}

func (db UserDB) List(ctx context.Context) ([]*otf.User, error) {
	selectBuilder := psql.Select("*").From("users")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var users []*otf.User
	if err := db.Select(&users, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan users from db: %w", err)
	}

	return users, nil
}

// Get retrieves a user from the DB, along with its sessions.
func (db UserDB) Get(ctx context.Context, spec otf.UserSpec) (*otf.User, error) {
	return getUser(ctx, db.DB, spec)
}

// CreateSession inserts the session, associating it with the user.
func (db UserDB) CreateSession(ctx context.Context, session *otf.Session) error {
	sql, args, err := db.BindNamed(insertSessionSQL, session)
	if err != nil {
		return err
	}

	_, err = db.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// UpdateSession updates a session row in the sessions table with the given
// session. The token identifies the session row to update.
func (db UserDB) UpdateSession(ctx context.Context, token string, updated *otf.Session) error {
	existing, err := getSession(ctx, db.DB, token)
	if err != nil {
		return err
	}

	updateBuilder := psql.
		Update("sessions").
		Where("token = ?", updated.Token)

	var modified bool

	if existing.Address != updated.Address {
		return fmt.Errorf("address cannot be updated on a session")
	}

	if existing.Flash != updated.Flash {
		modified = true
		updateBuilder = updateBuilder.Set("flash", updated.Flash)
	}

	if existing.Organization != updated.Organization {
		modified = true
		updateBuilder = updateBuilder.Set("organization", updated.Organization)
	}

	if existing.Expiry != updated.Expiry {
		modified = true
		updateBuilder = updateBuilder.Set("expiry", updated.Expiry)
	}

	if existing.UserID != updated.UserID {
		modified = true
		updateBuilder = updateBuilder.Set("user_id", updated.UserID)
	}

	if !modified {
		return fmt.Errorf("update was requested but no changes were found")
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

// Delete deletes a user from the DB.
func (db UserDB) Delete(ctx context.Context, spec otf.UserSpec) error {
	user, err := getUser(ctx, db.DB, spec)
	if err != nil {
		return err
	}

	sql, args, err := psql.Delete("users").Where("user_id = ?", user.ID).ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// DeleteSession deletes a user's session from the DB.
func (db UserDB) DeleteSession(ctx context.Context, token string) error {
	_, err := db.Exec("DELETE FROM sessions WHERE token = $1", token)
	if err != nil {
		return fmt.Errorf("unable to delete session: %w", err)
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

func (db UserDB) deleteExpired() error {
	_, err := db.Exec("DELETE FROM sessions WHERE expiry < current_timestamp")
	if err != nil {
		return fmt.Errorf("unable to delete expired sessions: %w", err)
	}

	return nil
}

func (db UserDB) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		db.deleteExpired()
	}
}

func getUser(ctx context.Context, db Getter, spec otf.UserSpec) (*otf.User, error) {
	selectBuilder := psql.
		Select(asColumnList("users", false, userColumns...)).
		From("users")

	switch {
	case spec.Username != nil:
		selectBuilder = selectBuilder.Where("username = ?", *spec.Username)
	case spec.SessionToken != nil:
		selectBuilder = selectBuilder.
			Join("sessions USING (user_id)").
			Where("sessions.token = ?", *spec.SessionToken)
	case spec.AuthenticationTokenID != nil:
		selectBuilder = selectBuilder.
			Join("tokens USING (token_id)").
			Where("tokens.token_id = ?", *spec.AuthenticationTokenID)
	case spec.AuthenticationToken != nil:
		selectBuilder = selectBuilder.
			Join("tokens USING (token)").
			Where("tokens.token = ?", *spec.AuthenticationToken)
	default:
		return nil, fmt.Errorf("empty user spec provided")
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building SQL query: %w", err)
	}

	// get user
	var user otf.User
	if err := db.Get(&user, sql, args...); err != nil {
		return nil, databaseError(err, sql)
	}

	// ...and their sessions
	user.Sessions, err = listSessions(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}

	// ...and their auth tokens
	user.Tokens, err = listTokens(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}

	// ...and their organizations
	user.Organizations, err = listOrganizationMemberships(ctx, db, user.ID)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// listSessions lists sessions belonging to the user with the given userID.
func listSessions(ctx context.Context, db Getter, userID string) ([]*otf.Session, error) {
	selectBuilder := psql.
		Select(sessionColumns...).
		From("sessions").
		Where("user_id = ?", userID).
		Where("expiry > current_timestamp")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var sessions []*otf.Session
	if err := db.Select(&sessions, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan sessions from db: %w", err)
	}

	return sessions, nil
}

// listTokens lists tokens belonging to the user with the given userID.
func listTokens(ctx context.Context, db Getter, userID string) ([]*otf.Token, error) {
	selectBuilder := psql.
		Select(tokenColumns...).
		From("tokens").
		Where("user_id = ?", userID)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var tokens []*otf.Token
	if err := db.Select(&tokens, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan tokens from db: %w", err)
	}

	return tokens, nil
}

func getSession(ctx context.Context, db Getter, token string) (*otf.Session, error) {
	selectBuilder := psql.
		Select(sessionColumns...).
		From("sessions").
		Where("token = ?", token).
		Where("expiry > current_timestamp")

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building SQL query: %w", err)
	}

	var session otf.Session
	if err := db.Get(&session, sql, args...); err != nil {
		return nil, databaseError(err, sql)
	}

	return &session, nil
}

// listOrganizationMemberships lists organizations belonging to the user with
// the given userID.
func listOrganizationMemberships(ctx context.Context, db Getter, userID string) ([]*otf.Organization, error) {
	selectBuilder := psql.
		Select(asColumnList("organizations", false, organizationColumns...)).
		From("organization_memberships").
		Join("organizations USING (organization_id)").
		Where("user_id = ?", userID)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var organizations []*otf.Organization
	if err := db.Select(&organizations, sql, args...); err != nil {
		return nil, fmt.Errorf("unable to scan organizations from db: %w", err)
	}

	return organizations, nil
}

func updateOrganizationMemberships(ctx context.Context, db *sqlx.DB, existing, updated *otf.User) error {
	added, removed := diffOrganizationLists(existing.Organizations, updated.Organizations)

	if err := addOrganizationMemberships(ctx, db, existing, added); err != nil {
		return err
	}

	if err := deleteOrganizationMemberships(ctx, db, existing, removed); err != nil {
		return err
	}

	return nil
}

func addOrganizationMemberships(ctx context.Context, db *sqlx.DB, user *otf.User, organizations []*otf.Organization) error {
	if len(organizations) == 0 {
		return nil
	}

	insertBuilder := psql.Insert("organization_memberships").Columns("user_id", "organization_id")

	for _, org := range organizations {
		insertBuilder = insertBuilder.Values(user.ID, org.ID)
	}

	sql, args, err := insertBuilder.ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

func deleteOrganizationMemberships(ctx context.Context, db *sqlx.DB, user *otf.User, organizations []*otf.Organization) error {
	if len(organizations) == 0 {
		return nil
	}

	var where squirrel.Or
	for _, org := range organizations {
		where = append(where, squirrel.Eq{"user_id": user.ID, "organization_id": org.ID})
	}

	sql, args, err := psql.Delete("organization_memberships").Where(where).ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.Exec(sql, args...)
	if err != nil {
		return databaseError(err, sql)
	}

	return nil
}

// diffOrganizationLists compares two lists of organizations, al and bl, and
// returns organizations present in b but absent in a, and organizations absent
// in b but present in a. Uses the organization ID for comparison.
func diffOrganizationLists(al, bl []*otf.Organization) (added, removed []*otf.Organization) {
	am := make(map[string]bool)
	for _, ax := range al {
		am[ax.ID] = true
	}

	bm := make(map[string]bool)
	for _, bx := range bl {
		bm[bx.ID] = true
	}

	// find those in a but not in b
	for _, ax := range al {
		if _, ok := bm[ax.ID]; !ok {
			removed = append(removed, ax)
		}
	}

	// find those in b but in in a
	for _, bx := range bl {
		if _, ok := am[bx.ID]; !ok {
			added = append(added, bx)
		}
	}

	return
}
