package variable

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

// pgdb is a database of variables on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (pdb *pgdb) createWorkspaceVariable(ctx context.Context, workspaceID resource.TfeID, v *Variable) error {
	return pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := pdb.Exec(ctx, `
INSERT INTO workspace_variables (
    variable_id,
    workspace_id
) VALUES (
    $1,
    $2
)
`, v.ID, workspaceID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (pdb *pgdb) listWorkspaceVariables(ctx context.Context, workspaceID resource.TfeID) ([]*Variable, error) {
	rows := pdb.Query(ctx, `
SELECT v.variable_id, v.key, v.value, v.description, v.category, v.sensitive, v.hcl, v.version_id
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE workspace_id = $1
`, workspaceID)
	return sql.CollectRows(rows, pgx.RowToAddrOfStructByName[Variable])
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error) {
	row := pdb.Query(ctx, `
SELECT
    workspace_id,
    v::"variables" AS variable
FROM workspace_variables
JOIN variables v USING (variable_id)
WHERE v.variable_id = $1
`, variableID)
	return sql.CollectOneRow(row, scanWorkspaceVariable)
}

func (pdb *pgdb) deleteWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error) {
	row := pdb.Query(ctx, `
DELETE
FROM workspace_variables wv USING variables v
WHERE wv.variable_id = $1
RETURNING wv.workspace_id, (v.*)::"variables" AS variable
`,
		variableID)
	return sql.CollectOneRow(row, scanWorkspaceVariable)
}

func scanWorkspaceVariable(row pgx.CollectableRow) (*WorkspaceVariable, error) {
	type model struct {
		*Variable
		WorkspaceID resource.TfeID `db:"workspace_id"`
	}
	m, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	wv := &WorkspaceVariable{
		Variable:    m.Variable,
		WorkspaceID: m.WorkspaceID,
	}
	return wv, nil
}

func (pdb *pgdb) createVariableSet(ctx context.Context, set *VariableSet) error {
	_, err := pdb.Exec(ctx, `
INSERT INTO variable_sets (
    variable_set_id,
    global,
    name,
    description,
    organization_name
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
`,
		set.ID,
		set.Global,
		set.Name,
		set.Description,
		set.Organization,
	)
	return err
}

func (pdb *pgdb) updateVariableSet(ctx context.Context, set *VariableSet) error {
	return pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := pdb.Exec(ctx, `
UPDATE variable_sets
SET
    global = $1,
    name = $2,
    description = $3
WHERE variable_set_id = $4
`,
			set.Global,
			set.Name,
			set.Description,
			set.ID,
		)
		if err != nil {
			return fmt.Errorf("updating variable set: %w", err)
		}
		// lazily delete all variable set workspaces, and then add them again,
		// regardless of whether there are any changes
		return pdb.Lock(ctx, "variable_set_workspaces", func(ctx context.Context, _ sql.Connection) error {
			if err := pdb.deleteAllVariableSetWorkspaces(ctx, set.ID); err != nil {
				if !errors.Is(err, internal.ErrResourceNotFound) {
					return err
				}
			}
			if err := pdb.createVariableSetWorkspaces(ctx, set.ID, set.Workspaces); err != nil {
				return err
			}
			return nil
		})
	})
}

func (pdb *pgdb) getVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error) {
	row := pdb.Query(ctx, `
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE vs.variable_set_id = $1
`,
		setID)
	return sql.CollectOneRow(row, scanVariableSet)
}

func (pdb *pgdb) getVariableSetByVariableID(ctx context.Context, variableID resource.TfeID) (*VariableSet, error) {
	row := pdb.Query(ctx, `
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_variables vsv USING (variable_set_id)
WHERE vsv.variable_id = $1
`, variableID)
	return sql.CollectOneRow(row, scanVariableSet)
}

func (pdb *pgdb) listVariableSets(ctx context.Context, organization organization.Name) ([]*VariableSet, error) {
	rows := pdb.Query(ctx, `
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
WHERE organization_name = $1
`, organization)
	return sql.CollectRows(rows, scanVariableSet)
}

func (pdb *pgdb) listVariableSetsByWorkspace(ctx context.Context, workspaceID resource.TfeID) ([]*VariableSet, error) {
	rows := pdb.Query(ctx, `
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN variable_set_workspaces vsw USING (variable_set_id)
WHERE vsw.workspace_id = $1
UNION
SELECT
    vs.variable_set_id, vs.global, vs.name, vs.description, vs.organization_name,
    (
        SELECT array_agg(v.*)::variables[]
        FROM variables v
        JOIN variable_set_variables vsv USING (variable_id)
        WHERE vsv.variable_set_id = vs.variable_set_id
    ) AS variables,
    (
        SELECT array_agg(vsw.workspace_id)::text[]
        FROM variable_set_workspaces vsw
        WHERE vsw.variable_set_id = vs.variable_set_id
    ) AS workspace_ids
FROM variable_sets vs
JOIN (organizations o JOIN workspaces w ON o.name = w.organization_name) ON o.name = vs.organization_name
WHERE vs.global IS true
AND w.workspace_id = $1
`, workspaceID)
	return sql.CollectRows(rows, scanVariableSet)
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID resource.TfeID) error {
	_, err := pdb.Exec(ctx, `
DELETE
FROM variable_sets
WHERE variable_set_id = $1
RETURNING variable_set_id, global, name, description, organization_name
`, setID)
	return err
}

func scanVariableSet(row pgx.CollectableRow) (*VariableSet, error) {
	type model struct {
		ID           resource.TfeID    `db:"variable_set_id"`
		Workspaces   []resource.TfeID  `db:"workspace_ids"`
		Organization organization.Name `db:"organization_name"`
		Name         string
		Description  string
		Global       bool
		Variables    []*Variable
	}
	m, err := pgx.RowToStructByName[model](row)
	if err != nil {
		return nil, err
	}
	vs := &VariableSet{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		Global:       m.Global,
		Variables:    m.Variables,
		Organization: m.Organization,
		Workspaces:   make([]resource.TfeID, len(m.Workspaces)),
	}
	copy(vs.Workspaces, m.Workspaces)
	return vs, nil
}

func (pdb *pgdb) addVariableToSet(ctx context.Context, setID resource.TfeID, v *Variable) error {
	return pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := pdb.Exec(ctx, `
INSERT INTO variable_set_variables (
    variable_set_id,
    variable_id
) VALUES (
    $1,
    $2
)`, setID, v.ID)
		if err != nil {
			return err
		}
		return nil
	})
}

func (pdb *pgdb) createVariableSetWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	err := pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		for _, wid := range workspaceIDs {
			_, err := pdb.Exec(ctx, `
INSERT INTO variable_set_workspaces (
    variable_set_id,
    workspace_id
) VALUES (
    $1,
    $2
)
`, setID, wid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (pdb *pgdb) deleteAllVariableSetWorkspaces(ctx context.Context, setID resource.TfeID) error {
	_, err := pdb.Exec(ctx, `
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
`, setID)
	return err
}

func (pdb *pgdb) deleteVariableSetWorkspaces(ctx context.Context, setID resource.TfeID, workspaceIDs []resource.TfeID) error {
	err := pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		for _, wid := range workspaceIDs {
			_, err := pdb.Exec(ctx, `
DELETE
FROM variable_set_workspaces
WHERE variable_set_id = $1
AND workspace_id = $2
`, setID, wid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

func (pdb *pgdb) createVariable(ctx context.Context, v *Variable) error {
	_, err := pdb.Exec(ctx, `
INSERT INTO variables (
    variable_id,
    key,
    value,
    description,
    category,
    sensitive,
    hcl,
    version_id
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
`,
		v.ID,
		v.Key,
		v.Value,
		v.Description,
		v.Category,
		v.Sensitive,
		v.HCL,
		v.VersionID,
	)
	return err
}

func (pdb *pgdb) updateVariable(ctx context.Context, v *Variable) error {
	_, err := pdb.Exec(ctx, `
UPDATE variables
SET
    key = $1,
    value = $2,
    description = $3,
    category = $4,
    sensitive = $5,
    hcl = $6,
    version_id = $7
WHERE variable_id = $8
`,
		v.Key,
		v.Value,
		v.Description,
		v.Category,
		v.Sensitive,
		v.HCL,
		v.VersionID,
		v.ID,
	)
	return err
}

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID resource.TfeID) error {
	_, err := pdb.Exec(ctx, ` DELETE FROM variables WHERE variable_id = $1 `, variableID)
	return err
}
