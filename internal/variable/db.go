package variable

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

// pgdb is a database of variables on postgres
type pgdb struct {
	*sql.DB // provides access to generated SQL queries
}

func (pdb *pgdb) create(ctx context.Context, v *Variable) error {
	_, err := pdb.Conn(ctx).InsertVariable(ctx, pggen.InsertVariableParams{
		VariableID:  sql.String(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   v.Sensitive,
		VersionID:   sql.String(v.VersionID),
		HCL:         v.HCL,
		WorkspaceID: sql.String(v.WorkspaceID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (pdb *pgdb) update(ctx context.Context, variableID string, fn func(*Variable) error) (*Variable, error) {
	var variable *Variable
	err := pdb.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
		// retrieve variable
		row, err := q.FindVariableForUpdate(ctx, sql.String(variableID))
		if err != nil {
			return err
		}
		variable = pgRow(row).toVariable()

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

func (pdb *pgdb) list(ctx context.Context, workspaceID string) ([]*Variable, error) {
	rows, err := pdb.Conn(ctx).FindVariables(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, sql.Error(err)
	}

	var variables []*Variable
	for _, row := range rows {
		variables = append(variables, pgRow(row).toVariable())
	}
	return variables, nil
}

func (pdb *pgdb) get(ctx context.Context, variableID string) (*Variable, error) {
	row, err := pdb.Conn(ctx).FindVariable(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return pgRow(row).toVariable(), nil
}

func (pdb *pgdb) delete(ctx context.Context, variableID string) (*Variable, error) {
	row, err := pdb.Conn(ctx).DeleteVariableByID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(row).toVariable(), nil
}

type pgRow struct {
	VariableID  pgtype.Text `json:"variable_id"`
	Key         pgtype.Text `json:"key"`
	Value       pgtype.Text `json:"value"`
	Description pgtype.Text `json:"description"`
	Category    pgtype.Text `json:"category"`
	Sensitive   bool        `json:"sensitive"`
	HCL         bool        `json:"hcl"`
	WorkspaceID pgtype.Text `json:"workspace_id"`
	VersionID   pgtype.Text `json:"version_id"`
}

func (row pgRow) toVariable() *Variable {
	return &Variable{
		ID:          row.VariableID.String,
		Key:         row.Key.String,
		Value:       row.Value.String,
		Description: row.Description.String,
		Category:    VariableCategory(row.Category.String),
		Sensitive:   row.Sensitive,
		HCL:         row.HCL,
		WorkspaceID: row.WorkspaceID.String,
		VersionID:   row.VersionID.String,
	}
}
