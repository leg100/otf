package variable

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

type (
	// pgdb is a database of variables on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	variableRow struct {
		VariableID  pgtype.Text `json:"variable_id"`
		Key         pgtype.Text `json:"key"`
		Value       pgtype.Text `json:"value"`
		Description pgtype.Text `json:"description"`
		Category    pgtype.Text `json:"category"`
		Sensitive   bool        `json:"sensitive"`
		HCL         bool        `json:"hcl"`
		VersionID   pgtype.Text `json:"version_id"`
	}

	variableSetRow struct {
		VariableSetID    pgtype.Text       `json:"variable_set_id"`
		Global           bool              `json:"global"`
		Name             pgtype.Text       `json:"name"`
		Description      pgtype.Text       `json:"description"`
		OrganizationName pgtype.Text       `json:"organization_name"`
		Variables        []pggen.Variables `json:"variables"`
		WorkspaceIds     []string          `json:"workspace_ids"`
	}
)

func (row variableRow) convert() *Variable {
	return &Variable{
		ID:          row.VariableID.String,
		Key:         row.Key.String,
		Value:       row.Value.String,
		Description: row.Description.String,
		Category:    VariableCategory(row.Category.String),
		Sensitive:   row.Sensitive,
		HCL:         row.HCL,
		VersionID:   row.VersionID.String,
	}
}

func (row variableSetRow) convert() *VariableSet {
	set := &VariableSet{
		ID:           row.VariableSetID.String,
		Global:       row.Global,
		Description:  row.Description.String,
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
	}
	set.Variables = make([]*Variable, len(row.Variables))
	for i, v := range row.Variables {
		set.Variables[i] = variableRow(v).convert()
	}
	set.Workspaces = row.WorkspaceIds
	return set
}

func (pdb *pgdb) createWorkspaceVariable(ctx context.Context, workspaceID string, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := q.InsertWorkspaceVariable(ctx, sql.String(v.ID), sql.String(workspaceID))
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) listWorkspaceVariables(ctx context.Context, workspaceID string) ([]*Variable, error) {
	rows, err := pdb.Conn(ctx).FindWorkspaceVariablesByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	variables := make([]*Variable, len(rows))
	for i, row := range rows {
		variables[i] = variableRow(row).convert()
	}
	return variables, nil
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	row, err := pdb.Conn(ctx).FindWorkspaceVariableByVariableID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: row.WorkspaceID.String,
		Variable:    variableRow(*row.Variable).convert(),
	}, nil
}

func (pdb *pgdb) deleteWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	row, err := pdb.Conn(ctx).DeleteWorkspaceVariableByID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: row.WorkspaceID.String,
		Variable:    variableRow(*row.Variable).convert(),
	}, nil
}

func (pdb *pgdb) createVariableSet(ctx context.Context, set *VariableSet) error {
	_, err := pdb.Conn(ctx).InsertVariableSet(ctx, pggen.InsertVariableSetParams{
		VariableSetID:    sql.String(set.ID),
		Name:             sql.String(set.Name),
		Description:      sql.String(set.Description),
		Global:           set.Global,
		OrganizationName: sql.String(set.Organization),
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateVariableSet(ctx context.Context, set *VariableSet) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		_, err := q.UpdateVariableSetByID(ctx, pggen.UpdateVariableSetByIDParams{
			Name:          sql.String(set.Name),
			Description:   sql.String(set.Description),
			Global:        set.Global,
			VariableSetID: sql.String(set.ID),
		})
		if err != nil {
			return err
		}

		// lazily delete all variable set workspaces, and then add them again,
		// regardless of whether there are any changes
		return pdb.Lock(ctx, "variable_set_workspaces", func(ctx context.Context, q pggen.Querier) error {
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

func (pdb *pgdb) getVariableSet(ctx context.Context, setID string) (*VariableSet, error) {
	row, err := pdb.Conn(ctx).FindVariableSetBySetID(ctx, sql.String(setID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return variableSetRow(row).convert(), nil
}

func (pdb *pgdb) getVariableSetByVariableID(ctx context.Context, variableID string) (*VariableSet, error) {
	row, err := pdb.Conn(ctx).FindVariableSetByVariableID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return variableSetRow(row).convert(), nil
}

func (pdb *pgdb) listVariableSets(ctx context.Context, organization string) ([]*VariableSet, error) {
	rows, err := pdb.Conn(ctx).FindVariableSetsByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		sets[i] = variableSetRow(row).convert()
	}
	return sets, nil
}

func (pdb *pgdb) listVariableSetsByWorkspace(ctx context.Context, workspaceID string) ([]*VariableSet, error) {
	rows, err := pdb.Conn(ctx).FindVariableSetsByWorkspace(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		sets[i] = variableSetRow(row).convert()
	}
	return sets, nil
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID string) error {
	_, err := pdb.Conn(ctx).DeleteVariableSetByID(ctx, sql.String(setID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (pdb *pgdb) addVariableToSet(ctx context.Context, setID string, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		_, err := q.InsertVariableSetVariable(ctx, sql.String(setID), sql.String(v.ID))
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) createVariableSetWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, wid := range workspaceIDs {
			_, err := pdb.Conn(ctx).InsertVariableSetWorkspace(ctx, sql.String(setID), sql.String(wid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (pdb *pgdb) deleteAllVariableSetWorkspaces(ctx context.Context, setID string) error {
	_, err := pdb.Conn(ctx).DeleteVariableSetWorkspaces(ctx, sql.String(setID))
	return sql.Error(err)
}

func (pdb *pgdb) deleteVariableSetWorkspaces(ctx context.Context, setID string, workspaceIDs []string) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		for _, wid := range workspaceIDs {
			_, err := pdb.Conn(ctx).DeleteVariableSetWorkspace(ctx, sql.String(setID), sql.String(wid))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (pdb *pgdb) createVariable(ctx context.Context, v *Variable) error {
	_, err := pdb.Conn(ctx).InsertVariable(ctx, pggen.InsertVariableParams{
		VariableID:  sql.String(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   v.Sensitive,
		VersionID:   sql.String(v.VersionID),
		HCL:         v.HCL,
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateVariable(ctx context.Context, v *Variable) error {
	_, err := pdb.Conn(ctx).UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
		VariableID:  sql.String(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   v.Sensitive,
		VersionID:   sql.String(v.VersionID),
		HCL:         v.HCL,
	})
	return sql.Error(err)
}

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID string) error {
	_, err := pdb.Conn(ctx).DeleteVariableByID(ctx, sql.String(variableID))
	return sql.Error(err)
}
