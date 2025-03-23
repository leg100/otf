package variable

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

var q = &Queries{}

type (
	// pgdb is a database of variables on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	VariableRow struct {
		VariableID  resource.TfeID
		Key         pgtype.Text
		Value       pgtype.Text
		Description pgtype.Text
		Category    pgtype.Text
		Sensitive   pgtype.Bool
		HCL         pgtype.Bool
		VersionID   pgtype.Text
	}

	VariableSetRow struct {
		VariableSetID    resource.TfeID
		Global           pgtype.Bool
		Name             pgtype.Text
		Description      pgtype.Text
		OrganizationName resource.OrganizationName
		Variables        []VariableModel
		WorkspaceIds     []pgtype.Text
	}
)

func (row VariableRow) convert() *Variable {
	return &Variable{
		ID:          row.VariableID,
		Key:         row.Key.String,
		Value:       row.Value.String,
		Description: row.Description.String,
		Category:    VariableCategory(row.Category.String),
		Sensitive:   row.Sensitive.Bool,
		HCL:         row.HCL.Bool,
		VersionID:   row.VersionID.String,
	}
}

func (row VariableSetRow) convert() (*VariableSet, error) {
	set := &VariableSet{
		ID:           row.VariableSetID,
		Global:       row.Global.Bool,
		Description:  row.Description.String,
		Name:         row.Name.String,
		Organization: row.OrganizationName,
	}
	set.Variables = make([]*Variable, len(row.Variables))
	for i, v := range row.Variables {
		set.Variables[i] = VariableRow(v).convert()
	}
	set.Workspaces = make([]resource.TfeID, len(row.WorkspaceIds))
	for i, wid := range row.WorkspaceIds {
		if err := set.Workspaces[i].Scan(wid.String); err != nil {
			return nil, err
		}
	}
	return set, nil
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
	rows, err := q.FindWorkspaceVariablesByWorkspaceID(ctx, pdb.Conn(ctx), workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}

	variables := make([]*Variable, len(rows))
	for i, row := range rows {
		variables[i] = VariableRow(row).convert()
	}
	return variables, nil
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error) {
	row, err := q.FindWorkspaceVariableByVariableID(ctx, pdb.Conn(ctx), variableID)
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: row.WorkspaceID,
		Variable:    VariableRow(row.VariableModel).convert(),
	}, nil
}

func (pdb *pgdb) deleteWorkspaceVariable(ctx context.Context, variableID resource.TfeID) (*WorkspaceVariable, error) {
	_, err := pdb.Exec(ctx, `
DELETE
FROM workspace_variables wv USING variables v
WHERE wv.variable_id = $1
RETURNING wv.workspace_id, (v.*)::"variables" AS variable
`,
		variableID)
	if err != nil {
		return nil, err
	}

	return &WorkspaceVariable{
		WorkspaceID: row.WorkspaceID,
		Variable:    VariableRow(row.VariableModel).convert(),
	}, nil
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
			return err
		}
		// lazily delete all variable set workspaces, and then add them again,
		// regardless of whether there are any changes
		return pdb.Lock(ctx, "variable_set_workspaces", func(ctx context.Context, _ sql.Connection) error {
			if err := pdb.deleteAllVariableSetWorkspaces(ctx, set.ID); err != nil {
				return err
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
	return sql.CollectOneRow(row, pdb.scanVariableSet)
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
	return sql.CollectOneRow(row, pdb.scanVariableSet)
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
	return sql.CollectRows(rows, pdb.scanVariableSet)
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
	return sql.CollectRows(rows, pdb.scanVariableSet)
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
	return sql.Error(err)
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
RETURNING variable_set_id, workspace_id
`, setID, wid)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
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
		v.VersionID,
		v.HCL,
	)
	return sql.Error(err)
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
    version_id = $6,
    hcl = $7
WHERE variable_id = $8
`,
		v.Key,
		v.Value,
		v.Description,
		v.Category,
		v.Sensitive,
		v.VersionID,
		v.HCL,
		v.ID,
	)
	return err
}

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID resource.TfeID) error {
	_, err := pdb.Exec(ctx, `
DELETE
FROM variables
WHERE variable_id = $1
RETURNING variable_id, key, value, description, category, sensitive, hcl, version_id
`, variableID)
	return err
}

func (pdb *pgdb) scanWorkspaceVariable(row pgx.CollectableRow) (*WorkspaceVariable, error) {
	var wv WorkspaceVariable
	err := row.Scan(
		&wv.WorkspaceID,
	)
	if err != nil {
		return nil, err
	}
	return &wv, nil
}

func (pdb *pgdb) scanVariableSet(row pgx.CollectableRow) (*VariableSet, error) {
	var vs VariableSet
	err := row.Scan(
		&vs.ID,
		&vs.Global,
		&vs.Name,
		&vs.Description,
		&vs.Organization,
		&vs.Variables,
		&vs.Workspaces,
	)
	if err != nil {
		return nil, err
	}
	return &vs, nil
}
