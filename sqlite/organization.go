package sqlite

import (
	"fmt"
	"strings"
	"time"

	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.OrganizationStore = (*OrganizationDB)(nil)

	insertOrganizationSql = `INSERT INTO organizations (
    created_at,
    updated_at,
    external_id,
    name,
    email,
    session_remember,
    session_timeout)
VALUES (
	:created_at,
    :updated_at,
    :external_id,
    :name,
    :email,
    :session_remember,
    :session_timeout)
`

	getOrganizationColumns = `
organizations.created_at           AS organizations.created_at
organizations.updated_at           AS organizations.updated_at
organizations.external_id          AS organizations.external_id
organizations.name                 AS organizations.name
organizations.email                AS organizations.email
organizations.session_remember     AS organizations.session_remember
organizations.session_timeout      AS organizations.session_timeout
`
)

// Organization models a row in a organizations table.
type OrganizationDB struct {
	*sqlx.DB
}

func NewOrganizationDB(db *sqlx.DB) *OrganizationDB {
	return &OrganizationDB{
		DB: db,
	}
}

// Create persists a Organization to the DB.
func (db OrganizationDB) Create(org *otf.Organization) (*otf.Organization, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	// Insert
	result, err := tx.NamedExec(insertOrganizationSql, org)
	if err != nil {
		return nil, err
	}
	org.Model.ID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return org, nil
}

// Update persists an updated Organization to the DB. The existing org is
// fetched from the DB, the supplied func is invoked on the org, and the updated
// org is persisted back to the DB.
func (db OrganizationDB) Update(name string, fn func(*otf.Organization) error) (*otf.Organization, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	org, err := getOrganization(tx, name)
	if err != nil {
		return nil, err
	}

	before := otf.Organization{}
	copier.Copy(&before, org)

	// Update obj using client-supplied fn
	if err := fn(org); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before, org)
	if len(updates) == 0 {
		return org, nil
	}

	org.UpdatedAt = time.Now()
	updates["updated_at"] = org.UpdatedAt

	var sql strings.Builder
	fmt.Fprintln(&sql, "UPDATE organizations")

	for k := range updates {
		fmt.Fprintln(&sql, "SET %s = :%[1]s", k)
	}

	fmt.Fprintf(&sql, "WHERE %s = :id", org.Model.ID)

	_, err = db.NamedExec(sql.String(), updates)
	if err != nil {
		return nil, err
	}

	return org, tx.Commit()
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	type listParams struct {
		Limit  int
		Offset int
	}

	params := listParams{}

	var sql strings.Builder
	fmt.Fprintln(&sql, "SELECT", getOrganizationColumns, "FROM", "organizations")

	if opts.PageSize > 0 {
		fmt.Fprintln(&sql, "LIMIT :limit")
		params.Limit = opts.PageSize
	}

	if opts.PageNumber > 0 {
		fmt.Fprintln(&sql, "OFFSET :limit")
		params.Offset = (opts.PageNumber - 1) * opts.PageSize
	}

	var result []otf.Organization
	if err := db.Select(&result, sql.String(), params); err != nil {
		return nil, err
	}

	// Convert from []otf.Organization to []*otf.Organization
	var items []*otf.Organization
	for _, r := range result {
		items = append(items, &r)
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, len(items)),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	return getOrganization(db.DB, name)
}

// Delete organization. TODO: delete dependencies, i.e. everything else too
func (db OrganizationDB) Delete(name string) error {
	_, err = db.MustExec("DELETE FROM organizations WHERE name = ?", name)
	return err
}

func getOrganization(getter Getter, name string) (*otf.Organization, error) {
	sql := fmt.Sprintln("SELECT", getOrganizationColumns, "FROM organizations WHERE name = ?")

	var org otf.Organization
	if err := getter.Get(&org, sql, name); err != nil {
		return nil, err
	}

	return &org, nil
}
