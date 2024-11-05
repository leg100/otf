package variable

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/sqlc"
)

type (
	// pgdb is a database of variables on postgres
	pgdb struct {
		*sql.DB // provides access to generated SQL queries
	}

	VariableRow struct {
		VariableID  pgtype.Text
		Key         pgtype.Text
		Value       pgtype.Text
		Description pgtype.Text
		Category    pgtype.Text
		Sensitive   pgtype.Bool
		HCL         pgtype.Bool
		VersionID   pgtype.Text
	}

	VariableSetRow struct {
		VariableSetID    pgtype.Text
		Global           pgtype.Bool
		Name             pgtype.Text
		Description      pgtype.Text
		OrganizationName pgtype.Text
		Variables        []sqlc.Variable
		WorkspaceIds     []pgtype.Text
	}
)

func (row VariableRow) convert() *Variable {
	return &Variable{
		ID:          resource.ParseID(row.VariableID.String),
		Key:         row.Key.String,
		Value:       row.Value.String,
		Description: row.Description.String,
		Category:    VariableCategory(row.Category.String),
		Sensitive:   row.Sensitive.Bool,
		HCL:         row.HCL.Bool,
		VersionID:   row.VersionID.String,
	}
}

func (row VariableSetRow) convert() *VariableSet {
	set := &VariableSet{
		ID:           resource.ParseID(row.VariableSetID.String),
		Global:       row.Global.Bool,
		Description:  row.Description.String,
		Name:         row.Name.String,
		Organization: row.OrganizationName.String,
	}
	set.Variables = make([]*Variable, len(row.Variables))
	for i, v := range row.Variables {
		set.Variables[i] = VariableRow(v).convert()
	}
	set.Workspaces = make([]resource.ID, len(row.WorkspaceIds))
	for i, wid := range row.WorkspaceIds {
		set.Workspaces[i] = resource.ParseID(wid.String)
	}
	return set
}

func (pdb *pgdb) createWorkspaceVariable(ctx context.Context, workspaceID resource.ID, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		return q.InsertWorkspaceVariable(ctx, sqlc.InsertWorkspaceVariableParams{
			VariableID:  sql.ID(v.ID),
			WorkspaceID: sql.ID(workspaceID),
		})
	})
	return sql.Error(err)
}

func (pdb *pgdb) listWorkspaceVariables(ctx context.Context, workspaceID resource.ID) ([]*Variable, error) {
	rows, err := pdb.Querier(ctx).FindWorkspaceVariablesByWorkspaceID(ctx, sql.ID(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	variables := make([]*Variable, len(rows))
	for i, row := range rows {
		variables[i] = VariableRow(row).convert()
	}
	return variables, nil
}

func (pdb *pgdb) getWorkspaceVariable(ctx context.Context, variableID resource.ID) (*WorkspaceVariable, error) {
	row, err := pdb.Querier(ctx).FindWorkspaceVariableByVariableID(ctx, sql.ID(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: resource.ParseID(row.WorkspaceID.String),
		Variable:    VariableRow(row.Variable).convert(),
	}, nil
}

func (pdb *pgdb) deleteWorkspaceVariable(ctx context.Context, variableID resource.ID) (*WorkspaceVariable, error) {
	row, err := pdb.Querier(ctx).DeleteWorkspaceVariableByID(ctx, sql.ID(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return &WorkspaceVariable{
		WorkspaceID: resource.ParseID(row.WorkspaceID.String),
		Variable:    VariableRow(row.Variable).convert(),
	}, nil
}

func (pdb *pgdb) createVariableSet(ctx context.Context, set *VariableSet) error {
	err := pdb.Querier(ctx).InsertVariableSet(ctx, sqlc.InsertVariableSetParams{
		VariableSetID:    sql.ID(set.ID),
		Name:             sql.String(set.Name),
		Description:      sql.String(set.Description),
		Global:           sql.Bool(set.Global),
		OrganizationName: sql.String(set.Organization),
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateVariableSet(ctx context.Context, set *VariableSet) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		_, err := q.UpdateVariableSetByID(ctx, sqlc.UpdateVariableSetByIDParams{
			Name:          sql.String(set.Name),
			Description:   sql.String(set.Description),
			Global:        sql.Bool(set.Global),
			VariableSetID: sql.ID(set.ID),
		})
		if err != nil {
			return err
		}

		// lazily delete all variable set workspaces, and then add them again,
		// regardless of whether there are any changes
		return pdb.Lock(ctx, "variable_set_workspaces", func(ctx context.Context, q *sqlc.Queries) error {
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

func (pdb *pgdb) getVariableSet(ctx context.Context, setID resource.ID) (*VariableSet, error) {
	row, err := pdb.Querier(ctx).FindVariableSetBySetID(ctx, sql.ID(setID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return VariableSetRow(row).convert(), nil
}

func (pdb *pgdb) getVariableSetByVariableID(ctx context.Context, variableID resource.ID) (*VariableSet, error) {
	row, err := pdb.Querier(ctx).FindVariableSetByVariableID(ctx, sql.ID(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return VariableSetRow(row).convert(), nil
}

func (pdb *pgdb) listVariableSets(ctx context.Context, organization string) ([]*VariableSet, error) {
	rows, err := pdb.Querier(ctx).FindVariableSetsByOrganization(ctx, sql.String(organization))
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		sets[i] = VariableSetRow(row).convert()
	}
	return sets, nil
}

func (pdb *pgdb) listVariableSetsByWorkspace(ctx context.Context, workspaceID resource.ID) ([]*VariableSet, error) {
	rows, err := pdb.Querier(ctx).FindVariableSetsByWorkspace(ctx, sql.ID(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	sets := make([]*VariableSet, len(rows))
	for i, row := range rows {
		sets[i] = VariableSetRow(row).convert()
	}
	return sets, nil
}

func (pdb *pgdb) deleteVariableSet(ctx context.Context, setID resource.ID) error {
	_, err := pdb.Querier(ctx).DeleteVariableSetByID(ctx, sql.ID(setID))
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (pdb *pgdb) addVariableToSet(ctx context.Context, setID resource.ID, v *Variable) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		if err := pdb.createVariable(ctx, v); err != nil {
			return err
		}
		err := q.InsertVariableSetVariable(ctx, sqlc.InsertVariableSetVariableParams{
			VariableSetID: sql.ID(setID),
			VariableID:    sql.ID(v.ID),
		})
		return err
	})
	return sql.Error(err)
}

func (pdb *pgdb) createVariableSetWorkspaces(ctx context.Context, setID resource.ID, workspaceIDs []resource.ID) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		for _, wid := range workspaceIDs {
			err := pdb.Querier(ctx).InsertVariableSetWorkspace(ctx, sqlc.InsertVariableSetWorkspaceParams{
				VariableSetID: sql.ID(setID),
				WorkspaceID:   sql.ID(wid),
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return sql.Error(err)
}

func (pdb *pgdb) deleteAllVariableSetWorkspaces(ctx context.Context, setID resource.ID) error {
	err := pdb.Querier(ctx).DeleteVariableSetWorkspaces(ctx, sql.ID(setID))
	return sql.Error(err)
}

func (pdb *pgdb) deleteVariableSetWorkspaces(ctx context.Context, setID resource.ID, workspaceIDs []resource.ID) error {
	err := pdb.Tx(ctx, func(ctx context.Context, q *sqlc.Queries) error {
		for _, wid := range workspaceIDs {
			_, err := pdb.Querier(ctx).DeleteVariableSetWorkspace(ctx, sqlc.DeleteVariableSetWorkspaceParams{
				VariableSetID: sql.ID(setID),
				WorkspaceID:   sql.ID(wid),
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
	err := pdb.Querier(ctx).InsertVariable(ctx, sqlc.InsertVariableParams{
		VariableID:  sql.ID(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   sql.Bool(v.Sensitive),
		VersionID:   sql.String(v.VersionID),
		HCL:         sql.Bool(v.HCL),
	})
	return sql.Error(err)
}

func (pdb *pgdb) updateVariable(ctx context.Context, v *Variable) error {
	_, err := pdb.Querier(ctx).UpdateVariableByID(ctx, sqlc.UpdateVariableByIDParams{
		VariableID:  sql.ID(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   sql.Bool(v.Sensitive),
		VersionID:   sql.String(v.VersionID),
		HCL:         sql.Bool(v.HCL),
	})
	return sql.Error(err)
}

func (pdb *pgdb) deleteVariable(ctx context.Context, variableID resource.ID) error {
	_, err := pdb.Querier(ctx).DeleteVariableByID(ctx, sql.ID(variableID))
	return sql.Error(err)
}
