package variable

import (
	"context"

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
	err := pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		return q.InsertWorkspaceVariable(ctx, pdb.Conn(ctx), InsertWorkspaceVariableParams{
			VariableID:  v.ID,
			WorkspaceID: workspaceID,
		})
	})
	return sql.Error(err)
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
	row, err := q.DeleteWorkspaceVariableByID(ctx, pdb.Conn(ctx), variableID)
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: row.WorkspaceID,
		Variable:    VariableRow(row.VariableModel).convert(),
	}, nil
}

func (pdb *pgdb) createVariableSet(ctx context.Context, set *VariableSet) error {
	err := q.InsertVariableSet(ctx, pdb.Conn(ctx), InsertVariableSetParams{
		VariableSetID:    set.ID,
		Name:             sql.String(set.Name),
		Description:      sql.String(set.Description),
		Global:           sql.Bool(set.Global),
		OrganizationName: set.Organization,
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateVariableSet(ctx context.Context, set *VariableSet) error {
	err := pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		_, err := q.UpdateVariableSetByID(ctx, pdb.Conn(ctx), UpdateVariableSetByIDParams{
			Name:          sql.String(set.Name),
			Description:   sql.String(set.Description),
			Global:        sql.Bool(set.Global),
			VariableSetID: set.ID,
		})
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
	return sql.Error(err)
}

func (pdb *pgdb) getVariableSet(ctx context.Context, setID resource.TfeID) (*VariableSet, error) {
	row, err := q.FindVariableSetBySetID(ctx, pdb.Conn(ctx), setID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return VariableSetRow(row).convert()
}

func (pdb *pgdb) getVariableSetByVariableID(ctx context.Context, variableID resource.TfeID) (*VariableSet, error) {
	row, err := q.FindVariableSetByVariableID(ctx, pdb.Conn(ctx), variableID)
	if err != nil {
		return nil, sql.Error(err)
	}
	return VariableSetRow(row).convert()
}

func (pdb *pgdb) listVariableSets(ctx context.Context, organization organization.Name) ([]*VariableSet, error) {
	rows, err := q.FindVariableSetsByOrganization(ctx, pdb.Conn(ctx), organization)
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		var err error
		sets[i], err = VariableSetRow(row).convert()
		if err != nil {
			return nil, err
		}
	}
	return sets, nil
}

func (pdb *pgdb) listVariableSetsByWorkspace(ctx context.Context, workspaceID resource.TfeID) ([]*VariableSet, error) {
	rows, err := q.FindVariableSetsByWorkspace(ctx, pdb.Conn(ctx), workspaceID)
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		var err error
		sets[i], err = VariableSetRow(row).convert()
		if err != nil {
			return nil, err
		}
	}
	return sets, nil
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID resource.TfeID) error {
	_, err := q.DeleteVariableSetByID(ctx, pdb.Conn(ctx), setID)
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (pdb *pgdb) addVariableToSet(ctx context.Context, setID resource.TfeID, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, conn sql.Connection) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		err := q.InsertVariableSetVariable(ctx, pdb.Conn(ctx), InsertVariableSetVariableParams{
			VariableSetID: setID,
			VariableID:    v.ID,
		})
		return err
	})
	return sql.Error(err)
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
			_, err := q.DeleteVariableSetWorkspace(ctx, pdb.Conn(ctx), DeleteVariableSetWorkspaceParams{
				VariableSetID: setID,
				WorkspaceID:   wid,
			})
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
