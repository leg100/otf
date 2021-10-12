package sqlite

import (
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"github.com/mitchellh/copystructure"
)

var (
	_ otf.OrganizationStore = (*OrganizationDB)(nil)

	organizationColumnsWithoutID = []string{
		"created_at",
		"updated_at",
		"external_id",
		"name",
		"email",
		"session_remember",
		"session_timeout",
	}
	organizationColumns = append(organizationColumnsWithoutID, "id")

	insertOrganizationSQL = fmt.Sprintf("INSERT INTO organizations (%s) VALUES (%s)",
		strings.Join(organizationColumnsWithoutID, ", "),
		strings.Join(otf.PrefixSlice(organizationColumnsWithoutID, ":"), ", "))
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
	result, err := db.NamedExec(insertOrganizationSQL, org)
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

	// Make a copy for comparison with the updated obj
	before, err := copystructure.Copy(org)
	if err != nil {
		return nil, err
	}

	// Update obj using client-supplied fn
	if err := fn(org); err != nil {
		return nil, err
	}

	updates := FindUpdates(db.Mapper, before.(*otf.Organization), org)
	if len(updates) == 0 {
		return org, nil
	}

	org.UpdatedAt = time.Now()
	updates["updated_at"] = org.UpdatedAt

	sql := sq.Update("organizations").Where("id = ?", org.Model.ID)

	query, args, err := sql.SetMap(updates).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("executing SQL statement: %s: %w", query, err)
	}

	return org, tx.Commit()
}

func (db OrganizationDB) List(opts otf.OrganizationListOptions) (*otf.OrganizationList, error) {
	selectBuilder := sq.Select(strings.Join(organizationColumns, ",")).
		From("organizations").
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, _, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var result []otf.Organization
	if err := db.Select(&result, sql); err != nil {
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
	result, err := db.Exec("DELETE FROM organizations WHERE name = ?", name)
	if err != nil {
		return err
	}
	_, err = result.RowsAffected()
	if err != nil {
		return err
	}
	return nil
}

func getOrganization(getter Getter, name string) (*otf.Organization, error) {
	selectBuilder := sq.Select(strings.Join(organizationColumns, ",")).
		From("organizations").
		Where("name = ?", name)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var org otf.Organization
	if err := getter.Get(&org, sql, args...); err != nil {
		return nil, err
	}

	return &org, nil
}
