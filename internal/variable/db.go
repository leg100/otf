package variable

import (
	"context"

	"github.com/jackc/pgtype"
	"github.com/leg100/otf"
	"github.com/leg100/otf/sql"
	"github.com/leg100/otf/sql/pggen"
)

// db is a database of variables
type db interface {
	otf.DB

	create(ctx context.Context, variable *Variable) error
	list(ctx context.Context, workspaceID string) ([]*Variable, error)
	get(ctx context.Context, variableID string) (*Variable, error)
	update(ctx context.Context, variableID string, updateFn func(*Variable) error) (*Variable, error)
	delete(ctx context.Context, variableID string) (*Variable, error)
	tx(context.Context, func(db) error) error
}

// pgdb is a database of variables on postgres
type pgdb struct {
	otf.DB // provides access to generated SQL queries
}

func newPGDB(db otf.DB) *pgdb {
	return &pgdb{db}
}

func (pdb *pgdb) create(ctx context.Context, v *Variable) error {
	_, err := pdb.InsertVariable(ctx, pggen.InsertVariableParams{
		VariableID:  sql.String(v.ID),
		Key:         sql.String(v.Key),
		Value:       sql.String(v.Value),
		Description: sql.String(v.Description),
		Category:    sql.String(string(v.Category)),
		Sensitive:   v.Sensitive,
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
	err := pdb.tx(ctx, func(tx db) error {
		// retrieve variable
		row, err := tx.FindVariableForUpdate(ctx, sql.String(variableID))
		if err != nil {
			return err
		}
		variable = pgRow(row).toVariable()

		// update variable
		if err := fn(variable); err != nil {
			return err
		}
		// persist variable
		_, err = tx.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  sql.String(variableID),
			Key:         sql.String(variable.Key),
			Value:       sql.String(variable.Value),
			Description: sql.String(variable.Description),
			Category:    sql.String(string(variable.Category)),
			Sensitive:   variable.Sensitive,
			HCL:         variable.HCL,
		})
		return err
	})
	return variable, sql.Error(err)
}

func (pdb *pgdb) list(ctx context.Context, workspaceID string) ([]*Variable, error) {
	rows, err := pdb.FindVariables(ctx, sql.String(workspaceID))
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
	row, err := pdb.FindVariable(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}

	return pgRow(row).toVariable(), nil
}

func (pdb *pgdb) delete(ctx context.Context, variableID string) (*Variable, error) {
	row, err := pdb.DeleteVariableByID(ctx, sql.String(variableID))
	if err != nil {
		return nil, sql.Error(err)
	}
	return pgRow(row).toVariable(), nil
}

// tx constructs a new pgdb within a transaction.
func (pdb *pgdb) tx(ctx context.Context, callback func(db) error) error {
	return pdb.Tx(ctx, func(tx otf.DB) error {
		return callback(newPGDB(tx))
	})
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
	}
}
