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

	workspaceVariableRow struct {
		VariableID  pgtype.Text `json:"variable_id"`
		WorkspaceID pgtype.Text `json:"workspace_id"`
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

func (row workspaceVariableRow) convert() *WorkspaceVariable {
	return &WorkspaceVariable{
		Variable: &Variable{
			ID:          row.VariableID.String,
			Key:         row.Key.String,
			Value:       row.Value.String,
			Description: row.Description.String,
			Category:    VariableCategory(row.Category.String),
			Sensitive:   row.Sensitive,
			HCL:         row.HCL,
			VersionID:   row.VersionID.String,
		},
		WorkspaceID: row.WorkspaceID.String,
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

func (pdb *pgdb) createWorkspaceVariable(ctx context.Context, v *WorkspaceVariable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		if err := pdb.createVariable(ctx, v.Variable); err != nil {
			return err
		}
		_, err := q.InsertWorkspaceVariable(ctx, sql.String(v.ID), sql.String(v.WorkspaceID))
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateWorkspaceVariable(ctx context.Context, variableID string, fn func(*WorkspaceVariable) error) (*WorkspaceVariable, error) {
	var variable *WorkspaceVariable
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// retrieve variable
		row, err := q.FindWorkspaceVariableForUpdate(ctx, sql.String(variableID))
		if err != nil {
			return err
		}
		variable = workspaceVariableRow(row).convert()

		// update variable
		if err := fn(variable); err != nil {
			return err
		}
		// persist variable
		_, err = q.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  sql.String(variableID),
			Key:         sql.String(variable.Key),
			Value:       sql.String(variable.Value),
			Description: sql.String(variable.Description),
			Category:    sql.String(string(variable.Category)),
			Sensitive:   variable.Sensitive,
			VersionID:   sql.String(variable.VersionID),
			HCL:         variable.HCL,
		})
		return err
	})
	return variable, sql.Error(err)
}

func (pdb *pgdb) listWorkspaceVariables(ctx context.Context, workspaceID string) ([]*WorkspaceVariable, error) {
	rows, err := pdb.Conn(ctx).FindWorkspaceVariablesByWorkspaceID(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	var variables []*WorkspaceVariable
	for _, row := range rows {
		variables = append(variables, workspaceVariableRow(row).convert())
	}
	return variables, nil
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID string) (*WorkspaceVariable, error) {
	row, err := pdb.Conn(ctx).FindWorkspaceVariableByID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return workspaceVariableRow(row).convert(), nil
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

func (pdb *pgdb) updateVariableSet(ctx context.Context, setID string, fn func(*VariableSet) error) (*VariableSet, error) {
	var set *VariableSet
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		row, err := q.FindVariableSetForUpdate(ctx, sql.String(setID))
		if err != nil {
			return err
		}
		set = variableSetRow(row).convert()

		// update variable set
		if err := fn(set); err != nil {
			return err
		}
		// persist updated variable set
		_, err = q.UpdateVariableSetByID(ctx, pggen.UpdateVariableSetByIDParams{
			Name:        sql.String(set.Name),
			Description: sql.String(set.Description),
			Global:      set.Global,
		})
		return err
	})
	return set, sql.Error(err)
}

func (pdb *pgdb) getVariableSet(ctx context.Context, setID string) (*VariableSet, error) {
	row, err := pdb.Conn(ctx).FindVariableSetByID(ctx, sql.String(setID))
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

	var sets []*VariableSet
	for _, row := range rows {
		sets = append(sets, variableSetRow(row).convert())
	}
	return sets, nil
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID string) error {
	_, err := pdb.Conn(ctx).DeleteVariableByID(ctx, sql.String(setID))
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

func (pdb *pgdb) updateVariableSetVariable(ctx context.Context, variableID string, fn func(*VariableSet) error) (*VariableSet, error) {
	var set *VariableSet
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// retrieve variable set along with the variable to update
		row, err := q.FindVariableSetVariableForUpdate(ctx, sql.String(variableID))
		if err != nil {
			return err
		}
		set = variableSetRow(row).convert()

		// update variable
		if err := fn(set); err != nil {
			return err
		}
		// persist variable
		_, err = q.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  sql.String(variableID),
			Key:         sql.String(set.Variables[0].Key),
			Value:       sql.String(set.Variables[0].Value),
			Description: sql.String(set.Variables[0].Description),
			Category:    sql.String(string(set.Variables[0].Category)),
			Sensitive:   set.Variables[0].Sensitive,
			VersionID:   sql.String(set.Variables[0].VersionID),
			HCL:         set.Variables[0].HCL,
		})
		return err
	})
	if err != nil {
		return nil, err
	}
	return set, nil
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

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID string) error {
	_, err := pdb.Conn(ctx).DeleteVariableByID(ctx, sql.String(variableID))
	return sql.Error(err)
}
