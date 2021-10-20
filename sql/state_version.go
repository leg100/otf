package sql

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
)

var (
	_ otf.StateVersionStore = (*StateVersionService)(nil)

	stateVersionColumns = []string{
		"state_version_id",
		"created_at",
		"updated_at",
		"serial",
		"blob_id",
	}

	stateVersionOutputColumns = []string{
		"state_version_output_id",
		"created_at",
		"updated_at",
		"name",
		"sensitive",
		"type",
		"value",
		"state_version_id",
	}

	insertStateVersionSQL = fmt.Sprintf("INSERT INTO state_versions (%s, workspace_id) VALUES (%s, :workspaces.workspace_id)",
		strings.Join(stateVersionColumns, ", "),
		strings.Join(otf.PrefixSlice(stateVersionColumns, ":"), ", "))

	insertStateVersionOutputSQL = fmt.Sprintf("INSERT INTO state_version_outputs (%s) VALUES (%s)",
		strings.Join(stateVersionOutputColumns, ", "),
		strings.Join(otf.PrefixSlice(stateVersionOutputColumns, ":"), ", "))
)

type StateVersionService struct {
	*sqlx.DB
}

func NewStateVersionDB(db *sqlx.DB) *StateVersionService {
	return &StateVersionService{
		DB: db,
	}
}

// Create persists a StateVersion to the DB.
func (s StateVersionService) Create(sv *otf.StateVersion) (*otf.StateVersion, error) {
	tx := s.MustBegin()
	defer tx.Rollback()

	// Insert state_version
	sql, args, err := tx.BindNamed(insertStateVersionSQL, sv)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(sql, args...)
	if err != nil {
		return nil, err
	}

	// Insert state_version_outputs
	for _, svo := range sv.Outputs {
		svo.StateVersionID = sv.ID
		sql, args, err := tx.BindNamed(insertStateVersionOutputSQL, svo)
		if err != nil {
			return nil, err
		}
		_, err = tx.Exec(sql, args...)
		if err != nil {
			return nil, err
		}
	}

	return sv, tx.Commit()
}

func (s StateVersionService) List(opts otf.StateVersionListOptions) (*otf.StateVersionList, error) {
	if opts.Workspace == nil {
		return nil, fmt.Errorf("missing required option: workspace")
	}
	if opts.Organization == nil {
		return nil, fmt.Errorf("missing required option: organization")
	}

	selectBuilder := psql.Select().From("state_versions").
		Join("workspaces USING (workspace_id)").
		Join("organizations USING (organization_id)").
		Where("workspaces.name = ?", *opts.Workspace).
		Where("organizations.name = ?", *opts.Organization)

	var count int
	if err := selectBuilder.Columns("count(*)").RunWith(s).QueryRow().Scan(&count); err != nil {
		return nil, fmt.Errorf("counting total rows: %w", err)
	}

	selectBuilder = selectBuilder.
		Columns(asColumnList("state_versions", false, stateVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		Limit(opts.GetLimit()).
		Offset(opts.GetOffset())

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var items []*otf.StateVersion
	if err := s.Select(&items, sql, args...); err != nil {
		return nil, err
	}

	return &otf.StateVersionList{
		Items:      items,
		Pagination: otf.NewPagination(opts.ListOptions, count),
	}, nil
}

func (s StateVersionService) Get(opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	return getStateVersion(s.DB, opts)
}

// Delete deletes a state version from the DB
func (s StateVersionService) Delete(id string) error {
	tx := s.MustBegin()
	defer tx.Rollback()

	sv, err := getStateVersion(tx, otf.StateVersionGetOptions{ID: otf.String(id)})
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM state_versions WHERE id = $1", sv.ID)
	if err != nil {
		return fmt.Errorf("unable to delete state_version: %w", err)
	}

	return tx.Commit()
}

func getStateVersion(getter Getter, opts otf.StateVersionGetOptions) (*otf.StateVersion, error) {
	selectBuilder := psql.Select(asColumnList("state_versions", false, stateVersionColumns...)).
		Columns(asColumnList("workspaces", true, workspaceColumns...)).
		From("state_versions").
		Join("workspaces USING (workspace_id)")

	switch {
	case opts.ID != nil:
		// Get state version by ID
		selectBuilder = selectBuilder.Where("state_versions.state_version_id = ?", *opts.ID)
	case opts.WorkspaceID != nil:
		// Get latest state version for given workspace
		selectBuilder = selectBuilder.Where("workspaces.workspace_id = ?", *opts.WorkspaceID)
		selectBuilder = selectBuilder.OrderBy("state_versions.serial DESC, state_versions.created_at DESC")
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	sv := otf.StateVersion{}
	if err := getter.Get(&sv, sql, args...); err != nil {
		return nil, databaseError(err)
	}

	if err := attachOutputs(getter, &sv); err != nil {
		return nil, err
	}

	return &sv, nil
}

func attachOutputs(getter Getter, sv *otf.StateVersion) error {
	selectBuilder := psql.Select("*").
		From("state_version_outputs").
		Where("state_version_id = ? ", sv.ID)

	sql, args, err := selectBuilder.ToSql()
	if err != nil {
		return err
	}

	outputs := []*otf.StateVersionOutput{}
	if err := getter.Select(&outputs, sql, args...); err != nil {
		return err
	}

	// Attach
	sv.Outputs = outputs

	return nil
}
