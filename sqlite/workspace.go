package sqlite

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/copier"
	"github.com/jmoiron/sqlx"
	"github.com/leg100/otf"
	"gorm.io/gorm"
)

var _ otf.WorkspaceStore = (*WorkspaceDB)(nil)

type WorkspaceDB struct {
	*sqlx.DB
}

func NewWorkspaceDB(db *sqlx.DB) *WorkspaceDB {
	return &WorkspaceDB{
		DB: db,
	}
}

// Create persists a Workspace to the DB. The returned Workspace is adorned with
// additional metadata, i.e. CreatedAt, UpdatedAt, etc.
func (db WorkspaceDB) Create(ws *otf.Workspace) (*otf.Workspace, error) {
	ws.CreatedAt = time.Now()
	ws.UpdatedAt = time.Now()

	clauses := map[string]interface{}{
		"created_at":         ws.CreatedAt,
		"updated_at":         ws.UpdatedAt,
		"external_id":        ws.ID,
		"allow_destroy_plan": ws.AllowDestroyPlan,
		"auto_apply":         ws.AutoApply,
		"locked":             ws.Locked,
		"organization_id":    ws.Organization.ID,
	}

	_, err := sq.Insert("ws").SetMap(clauses).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// Update persists an updated Workspace to the DB. The existing run is fetched
// from the DB, the supplied func is invoked on the run, and the updated run is
// persisted back to the DB. The returned Workspace includes any changes,
// including a new UpdatedAt value.
func (db WorkspaceDB) Update(spec otf.WorkspaceSpecifier, fn func(*otf.Workspace) error) (*otf.Workspace, error) {
	tx := db.MustBegin()
	defer tx.Rollback()

	ws, err := getWorkspace(tx, spec)
	if err != nil {
		return nil, err
	}

	before := otf.Workspace{}
	copier.Copy(&before, ws)

	// Update obj using client-supplied fn
	if err := fn(ws); err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	setIfChanged(before.AllowDestroyPlan, ws.AllowDestroyPlan, updates, "allow_destroy_plan")
	setIfChanged(before.AutoApply, ws.AutoApply, updates, "auto_apply")
	setIfChanged(before.ExecutionMode, ws.ExecutionMode, updates, "execution_mode")
	setIfChanged(before.Locked, ws.Locked, updates, "locked")

	if len(updates) == 0 {
		return ws, nil
	}

	updates["updated_at"] = time.Now()

	_, err = sq.Update("workspaces").SetMap(updates).Where("id = ?", ws.ID).RunWith(db).Exec()
	if err != nil {
		return nil, err
	}

	return ws, tx.Commit()
}

func (db WorkspaceDB) List(opts otf.WorkspaceListOptions) (*otf.WorkspaceList, error) {
	query := sq.Select("*").From("workspaces")

	if opts.OrganizationName != nil {
		org, err := getOrganizationByName(db.DB, *opts.OrganizationName)
		if err != nil {
			return nil, err
		}

		query = query.Where("organization_id = ?", org.ID)
	}

	if opts.Prefix != nil {
		query = query.Where("name LIKE ?", fmt.Sprintf("%s%%", *opts.Prefix))
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	wl := otf.WorkspaceList{}
	var count int

	rows, err := db.Queryx(sql, args)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		ws, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}

		wl.Items = append(wl.Items, ws)
		count++
	}

	wl.Pagination = otf.NewPagination(opts.ListOptions, count)

	return &wl, nil
}

func (db WorkspaceDB) Get(spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	return getWorkspace(db.MustBegin(), spec)
}

// Delete deletes a specific workspace, along with its associated records (runs
// etc).
func (db WorkspaceDB) Delete(spec otf.WorkspaceSpecifier) error {
	tx := db.MustBegin()
	defer tx.Rollback()

	ws, err := getWorkspace(tx, spec)
	if err != nil {
		return err
	}

	// Delete workspace
	_, err = sq.Delete("workspaces").Where("id = ?", ws.ID).RunWith(db).Exec()
	if err != nil {
		return err
	}

	// Delete associated runs
	_, err = sq.Delete("runs").Where("workspace_id = ?", ws.ID).RunWith(db).Exec()
	if err != nil {
		return err
	}

	// Delete associated state versions
	_, err = sq.Delete("state_versions").Where("workspace_id = ?", ws.ID).RunWith(db).Exec()
	if err != nil {
		return err
	}

	// Delete associated configuration versions
	_, err = sq.Delete("configuration_versions").Where("workspace_id = ?", ws.ID).RunWith(db).Exec()
	if err != nil {
		return err
	}

	return tx.Commit()
}

func getWorkspace(db sqlx.Queryer, spec otf.WorkspaceSpecifier) (*otf.Workspace, error) {
	query := sq.Select("workspaces.*").From("workspaces")

	switch {
	case spec.ID != nil:
		// Get workspace by (external) ID
		query = query.Where("external_id = ?", *spec.ID)
	case spec.InternalID != nil:
		// Get workspace by internal ID
		query = query.Where("id = ?", *spec.ID)
	case spec.Name != nil && spec.OrganizationName != nil:
		// Get workspace by name and organization name
		query = query.Join("JOIN organizations ON organizations.id = workspaces.organization_id").
			Where("workspaces.name = ? AND organizations.name = ?", spec.Name, spec.OrganizationName)
	default:
		return nil, otf.ErrInvalidWorkspaceSpecifier
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	ws, err := scanWorkspace(db.QueryRowx(sql, args))
	if err != nil {
		return nil, err
	}

	// Attach org to workspace
	ws.Organization, err = getOrganizationByID(db, ws.Organization.Model.ID)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func scanWorkspace(scannable StructScannable) (*otf.Workspace, error) {
	type result struct {
		metadata

		AllowDestroyPlan    bool `db:"allow_destroy_plan"`
		AutoApply           bool `db:"auto_apply"`
		CanQueueDestroyPlan bool `db:"can_queue_destroy_plan"`
		Description         string
		Environment         string
		ExecutionMode       string `db:"execution_mode"`
		FileTriggersEnabled bool   `db:"file_triggers_enabled"`
		Locked              bool
		Name                string
		QueueAllRuns        bool   `db:"queue_all_runs"`
		SpeculativeEnabled  bool   `db:"speculative_enabled"`
		SourceName          string `db:"source_name"`
		SourceURL           string `db:"source_url"`
		TerraformVersion    string `db:"terraform_version"`
		TriggerPrefixes     string `db:"trigger_prefixes"`
		WorkingDirectory    string `db:"working_directory"`

		OrganizationID uint `db:"organization_id"`
	}

	res := result{}
	if err := scannable.StructScan(res); err != nil {
		return nil, err
	}

	ws := otf.Workspace{
		ID: res.ExternalID,
		Model: gorm.Model{
			ID:        res.ID,
			CreatedAt: res.CreatedAt,
			UpdatedAt: res.UpdatedAt,
		},
		AllowDestroyPlan:    res.AllowDestroyPlan,
		AutoApply:           res.AutoApply,
		CanQueueDestroyPlan: res.CanQueueDestroyPlan,
		Description:         res.Description,
		Environment:         res.Environment,
		ExecutionMode:       res.ExecutionMode,
		FileTriggersEnabled: res.FileTriggersEnabled,
		Locked:              res.Locked,
		Name:                res.Name,
		QueueAllRuns:        res.QueueAllRuns,
		SpeculativeEnabled:  res.SpeculativeEnabled,
		SourceName:          res.SourceName,
		SourceURL:           res.SourceURL,
		TerraformVersion:    res.TerraformVersion,
		WorkingDirectory:    res.WorkingDirectory,
		Organization: &otf.Organization{
			Model: gorm.Model{
				ID: res.OrganizationID,
			},
		},
	}

	return &ws, nil
}
