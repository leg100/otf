package sql

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.UserStore = (*UserDB)(nil)

	DefaultSessionCleanupInterval = 5 * time.Minute

	sessionColumns = []string{
		"token",
		"created_at",
		"updated_at",
		"flash",
		"address",
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

	insertSessionSQL = `INSERT INTO sessions (token, flash, address, created_at, updated_at, expiry, user_id)
VALUES (:token, :flash, :address, :created_at, :updated_at, :expiry, :user_id)`

	insertTokenSQL = `INSERT INTO tokens (token_id, token, created_at, updated_at, description, user_id)
VALUES (:token_id, :token, :created_at, :updated_at, :description, :user_id)`
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

// CreateSession inserts the session, associating it with the user.
func (db UserDB) CreateSession(ctx context.Context, session *otf.Session) error {
	q := NewQuerier(db.Conn)

	_, err := q.InsertSession(ctx, InsertSessionParams{
		Token:   &session.Token,
		Address: &session.Address,
		Expiry:  session.Expiry,
		UserID:  &session.UserID,
	})
	return err
}

// UpdateSession updates a session row in the sessions table with the given
// session. The token identifies the session row to update.
func (db UserDB) UpdateSession(ctx context.Context, token string, updated *otf.Session) error {
	q := NewQuerier(db.Conn)

	existing, err := getSession(ctx, q, token)
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

func (db UserDB) UpdateOrganizationMemberships(ctx context.Context, id string, fn func(*otf.User, otf.OrganizationMembershipUpdater) error) (*otf.User, error) {
	tx, err := db.Conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	q := NewQuerier(tx)

	// select ...for update
	result, err := q.FindUserByIDForUpdate(ctx, &id)
	if err != nil {
		return nil, err
	}
	user := convertUser(result)

	if err := fn(user, newUserStatusUpdater(tx, user.ID)); err != nil {
		return nil, err
	}

	return user, tx.Commit(ctx)
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
	q := NewQuerier(db.Conn)

	_, err := q.DeleteSessionsExpired(context.Background())
	return err
}

func (db UserDB) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		<-ticker.C
		db.deleteExpired()
	}
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
func convertUser(row userRow) *otf.User {
	user := otf.User{
		ID:                  *row.GetUserID(),
		Timestamps:          convertTimestamps(row),
		Username:            *row.GetUsername(),
		CurrentOrganization: row.GetCurrentOrganization(),
	}

	return &user
}
