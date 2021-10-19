package sql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.OrganizationStore = (*OrganizationDB)(nil)

	organizationColumns = []string{
		"organization_id",
		"created_at",
		"updated_at",
		"name",
		"email",
		"session_remember",
		"session_timeout",
	}

	insertOrganizationSQL = fmt.Sprintf("INSERT INTO organizations (%s) VALUES (%s)",
		strings.Join(organizationColumns, ", "),
		strings.Join(otf.PrefixSlice(organizationColumns, ":"), ", "))
)

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
	// Insert
	sql, args, err := db.BindNamed(insertOrganizationSQL, org)
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(sql, args...)
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

	// Make a copy for comparison with the updated obj
	before, err := copystructure.Copy(org)
	if err != nil {
		return nil, err
	}

	// Update obj using client-supplied fn
	if err := fn(org); err != nil {
		return nil, err
	}

	updated, err := update(db.Mapper, tx, "organizations", "organization_id", before.(*otf.Organization), org)
	if err != nil {
		return nil, err
	}

	if updated {
		return org, tx.Commit()
	}

	return org, nil
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	selectBuilder := psql.Select().From("organizations")

	var count int
	if err := selectBuilder.Columns("count(*)").RunWith(db).QueryRow().Scan(&count); err != nil {
		return nil, fmt.Errorf("counting total rows: %w", err)
	}

	selectBuilder = selectBuilder.
		Columns(strings.Join(organizationColumns, ",")).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*otf.Organization
	if err := db.Select(&items, sql); err != nil {
		return nil, err
	}

	return &otf.OrganizationList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

func (db OrganizationDB) Get(name string) (*otf.Organization, error) {
	return getOrganization(db.DB, name)
}

func (db OrganizationDB) Delete(name string) error {
	result, err := db.Exec("DELETE FROM organizations WHERE name = $1", name)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return otf.ErrResourceNotFound
	}
	return nil
}

func getOrganization(getter Getter, name string) (*otf.Organization, error) {
	selectBuilder := psql.Select(strings.Join(organizationColumns, ",")).
		From("organizations").
		Where("name = ?", name)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var org otf.Organization
	if err := getter.Get(&org, sql, args...); err != nil {
		return nil, databaseError(err)
	}

	return &org, nil
}
