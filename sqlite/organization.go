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

	organizationsTableName = "organizations"

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

	organizationColumns = []string{
		"id",
		"created_at",
		"updated_at",
		"external_id",
		"name",
		"email",
		"session_remember",
		"session_timeout",
	}

	organizationColumnList = asColumnList(organizationsTableName, organizationColumns...)

	listOrganizationsSql = fmt.Sprintf(`SELECT %s FROM organizations LIMIT :limit OFFSET :offset`,
		asColumnList(organizationsTableName, organizationColumns...))

	getOrganizationSql = fmt.Sprintf("SELECT %s FROM organizations WHERE name = ?", organizationColumns)
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
	limit, offset := opts.GetSQLWindow()

	params := map[string]interface{}{
		"limit":  limit,
		"offset": offset,
	}

	var result []otf.Organization
	if err := db.Select(&result, listOrganizationsSql, params); err != nil {
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
	_, err := db.MustExec("DELETE FROM organizations WHERE name = ?", name)
	return err
}

func getOrganization(getter Getter, name string) (*otf.Organization, error) {
	var org otf.Organization
	if err := getter.Get(&org, getOrganizationSql, name); err != nil {
		return nil, err
	}

	return &org, nil
}
