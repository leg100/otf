package policy

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/vcs"
)

type pgdb struct {
	*sql.DB
}

func (db *pgdb) createPolicySet(ctx context.Context, set *PolicySet) error {
	_, err := db.Exec(ctx, `
INSERT INTO policy_sets (
	policy_set_id,
	created_at,
	updated_at,
	organization_name,
	name,
	description,
	source,
	vcs_provider_id,
	vcs_repo,
	vcs_ref,
	vcs_path,
	vcs_policy_paths,
	last_synced_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
`,
		set.ID,
		set.CreatedAt,
		set.UpdatedAt,
		set.Organization,
		set.Name,
		set.Description,
		set.Source,
		set.VCSProviderID,
		set.VCSRepo,
		set.VCSRef,
		set.VCSPath,
		set.VCSPolicyPaths,
		set.LastSyncedAt,
	)
	return err
}

func (db *pgdb) listPolicySets(ctx context.Context, org organization.Name) ([]*PolicySet, error) {
	rows := db.Query(ctx, `
SELECT policy_set_id, created_at, updated_at, organization_name, name, description, source, vcs_provider_id, vcs_repo, vcs_ref, vcs_path, vcs_policy_paths, last_synced_at
FROM policy_sets
WHERE organization_name = $1
ORDER BY name ASC
`, org)
	return sql.CollectRows(rows, db.scanPolicySet)
}

func (db *pgdb) listVCSPolicySetsByRepo(ctx context.Context, providerID resource.TfeID, repo vcs.Repo) ([]*PolicySet, error) {
	rows := db.Query(ctx, `
SELECT policy_set_id, created_at, updated_at, organization_name, name, description, source, vcs_provider_id, vcs_repo, vcs_ref, vcs_path, vcs_policy_paths, last_synced_at
FROM policy_sets
WHERE source = $1
AND vcs_provider_id = $2
AND vcs_repo = $3
ORDER BY name ASC
`, VCSPolicySetSource, providerID, repo.String())
	return sql.CollectRows(rows, db.scanPolicySet)
}

func (db *pgdb) getPolicySet(ctx context.Context, id resource.TfeID) (*PolicySet, error) {
	rows := db.Query(ctx, `
SELECT policy_set_id, created_at, updated_at, organization_name, name, description, source, vcs_provider_id, vcs_repo, vcs_ref, vcs_path, vcs_policy_paths, last_synced_at
FROM policy_sets
WHERE policy_set_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scanPolicySet)
}

func (db *pgdb) getPolicySetOrganization(ctx context.Context, id resource.TfeID) (organization.Name, error) {
	rows := db.Query(ctx, `SELECT organization_name FROM policy_sets WHERE policy_set_id = $1`, id)
	return sql.CollectOneType[organization.Name](rows)
}

func (db *pgdb) updatePolicySet(ctx context.Context, id resource.TfeID, fn func(context.Context, *PolicySet) error) (*PolicySet, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*PolicySet, error) {
			rows := db.Query(ctx, `
SELECT policy_set_id, created_at, updated_at, organization_name, name, description, source, vcs_provider_id, vcs_repo, vcs_ref, vcs_path, vcs_policy_paths, last_synced_at
FROM policy_sets
WHERE policy_set_id = $1
FOR UPDATE
`, id)
			return sql.CollectOneRow(rows, db.scanPolicySet)
		},
		fn,
		func(ctx context.Context, set *PolicySet) error {
			_, err := db.Exec(ctx, `
UPDATE policy_sets
SET name = $1,
	description = $2,
	source = $3,
	vcs_provider_id = $4,
	vcs_repo = $5,
	vcs_ref = $6,
	vcs_path = $7,
	vcs_policy_paths = $8,
	last_synced_at = $9,
	updated_at = $10
WHERE policy_set_id = $11
`, set.Name, set.Description, set.Source, set.VCSProviderID, set.VCSRepo, set.VCSRef, set.VCSPath, set.VCSPolicyPaths, set.LastSyncedAt, set.UpdatedAt, set.ID)
			return err
		},
	)
}

func (db *pgdb) deletePolicySet(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `DELETE FROM policy_sets WHERE policy_set_id = $1`, id)
	return err
}

func (db *pgdb) createPolicy(ctx context.Context, policy *Policy) error {
	_, err := db.Exec(ctx, `
INSERT INTO policies (
	policy_id,
	policy_set_id,
	created_at,
	updated_at,
	name,
	description,
	enforcement_level,
	source,
	path
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
		policy.ID,
		policy.PolicySetID,
		policy.CreatedAt,
		policy.UpdatedAt,
		policy.Name,
		policy.Description,
		policy.EnforcementLevel,
		policy.Source,
		policy.Path,
	)
	return err
}

func (db *pgdb) listPolicies(ctx context.Context, setID resource.TfeID) ([]*Policy, error) {
	rows := db.Query(ctx, `
SELECT
	p.policy_id,
	p.policy_set_id,
	p.created_at,
	p.updated_at,
	p.name,
	p.description,
	p.enforcement_level,
	p.source,
	p.path,
	ps.organization_name,
	ps.name AS policy_set_name
FROM policies p
JOIN policy_sets ps USING (policy_set_id)
WHERE p.policy_set_id = $1
ORDER BY p.name ASC
`, setID)
	return sql.CollectRows(rows, db.scanPolicy)
}

func (db *pgdb) listPoliciesByOrganization(ctx context.Context, org organization.Name) ([]*Policy, error) {
	rows := db.Query(ctx, `
SELECT
	p.policy_id,
	p.policy_set_id,
	p.created_at,
	p.updated_at,
	p.name,
	p.description,
	p.enforcement_level,
	p.source,
	p.path,
	ps.organization_name,
	ps.name AS policy_set_name
FROM policy_sets ps
JOIN policies p ON p.policy_set_id = ps.policy_set_id
WHERE ps.organization_name = $1
ORDER BY ps.name ASC, p.name ASC
`, org)
	return sql.CollectRows(rows, db.scanPolicy)
}

func (db *pgdb) listPoliciesByWorkspace(ctx context.Context, workspaceID resource.TfeID) ([]*Policy, error) {
	rows := db.Query(ctx, `
SELECT
	p.policy_id,
	p.policy_set_id,
	p.created_at,
	p.updated_at,
	p.name,
	p.description,
	p.enforcement_level,
	p.source,
	p.path,
	ps.organization_name,
	ps.name AS policy_set_name
FROM policy_set_workspaces psw
JOIN policy_sets ps ON ps.policy_set_id = psw.policy_set_id
JOIN policies p ON p.policy_set_id = ps.policy_set_id
WHERE psw.workspace_id = $1
ORDER BY ps.name ASC, p.name ASC
`, workspaceID)
	return sql.CollectRows(rows, db.scanPolicy)
}

func (db *pgdb) getPolicy(ctx context.Context, id resource.TfeID) (*Policy, error) {
	rows := db.Query(ctx, `
SELECT
	p.policy_id,
	p.policy_set_id,
	p.created_at,
	p.updated_at,
	p.name,
	p.description,
	p.enforcement_level,
	p.source,
	p.path,
	ps.organization_name,
	ps.name AS policy_set_name
FROM policies p
JOIN policy_sets ps USING (policy_set_id)
WHERE p.policy_id = $1
`, id)
	return sql.CollectOneRow(rows, db.scanPolicy)
}

func (db *pgdb) getPolicySetIDByPolicy(ctx context.Context, id resource.TfeID) (resource.TfeID, error) {
	rows := db.Query(ctx, `SELECT policy_set_id FROM policies WHERE policy_id = $1`, id)
	return sql.CollectOneType[resource.TfeID](rows)
}

func (db *pgdb) updatePolicy(ctx context.Context, id resource.TfeID, fn func(context.Context, *Policy) error) (*Policy, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*Policy, error) {
			rows := db.Query(ctx, `
SELECT
	p.policy_id,
	p.policy_set_id,
	p.created_at,
	p.updated_at,
	p.name,
	p.description,
	p.enforcement_level,
	p.source,
	p.path,
	ps.organization_name,
	ps.name AS policy_set_name
FROM policies p
JOIN policy_sets ps USING (policy_set_id)
WHERE p.policy_id = $1
FOR UPDATE OF p
`, id)
			return sql.CollectOneRow(rows, db.scanPolicy)
		},
		fn,
		func(ctx context.Context, policy *Policy) error {
			_, err := db.Exec(ctx, `
UPDATE policies
SET name = $1,
	description = $2,
	enforcement_level = $3,
	source = $4,
	path = $5,
	updated_at = $6
WHERE policy_id = $7
`, policy.Name, policy.Description, policy.EnforcementLevel, policy.Source, policy.Path, policy.UpdatedAt, policy.ID)
			return err
		},
	)
}

func (db *pgdb) deletePolicy(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `DELETE FROM policies WHERE policy_id = $1`, id)
	return err
}

func (db *pgdb) createPolicyModule(ctx context.Context, mod *PolicyModule) error {
	_, err := db.Exec(ctx, `
INSERT INTO policy_modules (
	policy_module_id,
	policy_set_id,
	created_at,
	updated_at,
	name,
	path,
	source
) VALUES ($1, $2, $3, $4, $5, $6, $7)
`, mod.ID, mod.PolicySetID, mod.CreatedAt, mod.UpdatedAt, mod.Name, mod.Path, mod.Source)
	return err
}

func (db *pgdb) listPolicyModules(ctx context.Context, setID resource.TfeID) ([]*PolicyModule, error) {
	rows := db.Query(ctx, `
SELECT policy_module_id, policy_set_id, created_at, updated_at, name, path, source
FROM policy_modules
WHERE policy_set_id = $1
ORDER BY path ASC
`, setID)
	return sql.CollectRows(rows, db.scanPolicyModule)
}

func (db *pgdb) listPolicyModulesByOrganization(ctx context.Context, org organization.Name) ([]*PolicyModule, error) {
	rows := db.Query(ctx, `
SELECT pm.policy_module_id, pm.policy_set_id, pm.created_at, pm.updated_at, pm.name, pm.path, pm.source
FROM policy_sets ps
JOIN policy_modules pm USING (policy_set_id)
WHERE ps.organization_name = $1
ORDER BY pm.path ASC
`, org)
	return sql.CollectRows(rows, db.scanPolicyModule)
}

func (db *pgdb) updatePolicyModule(ctx context.Context, id resource.TfeID, fn func(context.Context, *PolicyModule) error) (*PolicyModule, error) {
	return sql.Updater(
		ctx,
		db.DB,
		func(ctx context.Context) (*PolicyModule, error) {
			rows := db.Query(ctx, `
SELECT policy_module_id, policy_set_id, created_at, updated_at, name, path, source
FROM policy_modules
WHERE policy_module_id = $1
FOR UPDATE
`, id)
			return sql.CollectOneRow(rows, db.scanPolicyModule)
		},
		fn,
		func(ctx context.Context, mod *PolicyModule) error {
			_, err := db.Exec(ctx, `
UPDATE policy_modules
SET name = $1,
	path = $2,
	source = $3,
	updated_at = $4
WHERE policy_module_id = $5
`, mod.Name, mod.Path, mod.Source, mod.UpdatedAt, mod.ID)
			return err
		},
	)
}

func (db *pgdb) deletePolicyModule(ctx context.Context, id resource.TfeID) error {
	_, err := db.Exec(ctx, `DELETE FROM policy_modules WHERE policy_module_id = $1`, id)
	return err
}

func (db *pgdb) setPolicySetWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	if _, err := db.Exec(ctx, `DELETE FROM policy_set_workspaces WHERE policy_set_id = $1`, setID); err != nil {
		return err
	}
	for _, workspaceID := range workspaceIDs {
		if _, err := db.Exec(ctx, `
INSERT INTO policy_set_workspaces (policy_set_id, workspace_id) VALUES ($1, $2)
`, setID, workspaceID); err != nil {
			return err
		}
	}
	return nil
}

func (db *pgdb) listPolicySetWorkspaceIDs(ctx context.Context, setID resource.TfeID) ([]resource.TfeID, error) {
	rows := db.Query(ctx, `
SELECT workspace_id
FROM policy_set_workspaces
WHERE policy_set_id = $1
ORDER BY workspace_id
`, setID)
	return sql.CollectRows(rows, pgx.RowTo[resource.TfeID])
}

func (db *pgdb) replacePolicyChecks(ctx context.Context, runID resource.TfeID, checks []*PolicyCheck) error {
	if _, err := db.Exec(ctx, `DELETE FROM policy_checks WHERE run_id = $1`, runID); err != nil && !errors.Is(err, internal.ErrResourceNotFound) {
		return err
	}
	for _, check := range checks {
		_, err := db.Exec(ctx, `
INSERT INTO policy_checks (
	policy_check_id,
	run_id,
	workspace_id,
	policy_set_id,
	policy_id,
	organization_name,
	policy_name,
	policy_set_name,
	enforcement_level,
	passed,
	output,
	created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
`,
			check.ID,
			check.RunID,
			check.WorkspaceID,
			check.PolicySetID,
			check.PolicyID,
			check.Organization,
			check.PolicyName,
			check.PolicySetName,
			check.EnforcementLevel,
			check.Passed,
			check.Output,
			check.CreatedAt,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *pgdb) listPolicyChecks(ctx context.Context, runID resource.TfeID) ([]*PolicyCheck, error) {
	rows := db.Query(ctx, `
SELECT policy_check_id, run_id, workspace_id, policy_set_id, policy_id, organization_name, policy_name, policy_set_name, enforcement_level, passed, output, created_at
FROM policy_checks
WHERE run_id = $1
ORDER BY policy_set_name, policy_name
`, runID)
	return sql.CollectRows(rows, db.scanPolicyCheck)
}

func (db *pgdb) getRunIDByPolicyCheck(ctx context.Context, id resource.TfeID) (resource.TfeID, error) {
	rows := db.Query(ctx, `SELECT run_id FROM policy_checks WHERE policy_check_id = $1`, id)
	return sql.CollectOneType[resource.TfeID](rows)
}

func (db *pgdb) countWorkspaceFailures(ctx context.Context, workspaceID resource.TfeID) (int, error) {
	count, err := db.Int(ctx, `
SELECT count(*)
FROM policy_checks
WHERE workspace_id = $1
AND passed = false
AND enforcement_level = $2
`, workspaceID, MandatoryEnforcement)
	return int(count), err
}

func (db *pgdb) scanPolicySet(row pgx.CollectableRow) (*PolicySet, error) {
	var (
		set          PolicySet
		vcsProvider  *resource.TfeID
		vcsRepo      *string
		vcsRef       string
		vcsPath      string
		vcsPolicies  []string
		lastSyncedAt *time.Time
	)
	err := row.Scan(
		&set.ID,
		&set.CreatedAt,
		&set.UpdatedAt,
		&set.Organization,
		&set.Name,
		&set.Description,
		&set.Source,
		&vcsProvider,
		&vcsRepo,
		&vcsRef,
		&vcsPath,
		&vcsPolicies,
		&lastSyncedAt,
	)
	if err != nil {
		return nil, err
	}
	set.VCSProviderID = vcsProvider
	set.VCSRef = vcsRef
	set.VCSPath = vcsPath
	set.VCSPolicyPaths = vcsPolicies
	set.LastSyncedAt = lastSyncedAt
	if vcsRepo != nil && *vcsRepo != "" {
		repo, err := vcs.NewRepoFromString(*vcsRepo)
		if err != nil {
			return nil, err
		}
		set.VCSRepo = &repo
	}
	return &set, nil
}

func (db *pgdb) scanPolicy(row pgx.CollectableRow) (*Policy, error) {
	return pgx.RowToAddrOfStructByName[Policy](row)
}

func (db *pgdb) scanPolicyModule(row pgx.CollectableRow) (*PolicyModule, error) {
	return pgx.RowToAddrOfStructByName[PolicyModule](row)
}

func (db *pgdb) scanPolicyCheck(row pgx.CollectableRow) (*PolicyCheck, error) {
	return pgx.RowToAddrOfStructByName[PolicyCheck](row)
}
