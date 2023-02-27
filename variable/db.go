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
	otf.Database

	create(ctx context.Context, variable *Variable) error
	list(ctx context.Context, workspaceID string) ([]*Variable, error)
	get(ctx context.Context, variableID string) (*Variable, error)
	update(ctx context.Context, variableID string, updateFn func(*Variable) error) (*Variable, error)
	delete(ctx context.Context, variableID string) (*Variable, error)
	tx(context.Context, func(db) error) error
}

// pgdb is a database of variables on postgres
type pgdb struct {
	otf.Database // provides access to generated SQL queries
}

func newPGDB(db otf.Database) *pgdb {
	return &pgdb{db}
}

func (pdb *pgdb) create(ctx context.Context, v *Variable) error {
	_, err := pdb.InsertVariable(ctx, pggen.InsertVariableParams{
		VariableID:  sql.String(v.id),
		Key:         sql.String(v.key),
		Value:       sql.String(v.value),
		Description: sql.String(v.description),
		Category:    sql.String(string(v.category)),
		Sensitive:   v.sensitive,
		HCL:         v.hcl,
		WorkspaceID: sql.String(v.workspaceID),
	})
	if err != nil {
		return sql.Error(err)
	}
	return nil
}

func (pdb *pgdb) update(ctx context.Context, variableID string, fn func(*Variable) error) (*Variable, error) {
	var variable *Variable
	err := pdb.tx(ctx, func(tx db) error {
		var err error
		// retrieve variable
		row, err := tx.FindVariableForUpdate(ctx, sql.String(variableID))
		if err != nil {
			return sql.Error(err)
		}
		variable = pgRow(row).toVariable()

		// update variable
		if err := fn(variable); err != nil {
			return err
		}
		// persist variable
		_, err = tx.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  sql.String(variableID),
			Key:         sql.String(variable.Key()),
			Value:       sql.String(variable.Value()),
			Description: sql.String(variable.Description()),
			Category:    sql.String(string(variable.Category())),
			Sensitive:   variable.Sensitive(),
			HCL:         variable.HCL(),
		})
		return err
	})
	return variable, err
}

func (pdb *pgdb) list(ctx context.Context, workspaceID string) ([]*Variable, error) {
	rows, err := pdb.FindVariables(ctx, sql.String(workspaceID))
	if err != nil {
		return nil, err
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
		return nil, err
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
	return pdb.Transaction(ctx, func(tx otf.Database) error {
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
		id:          row.VariableID.String,
		key:         row.Key.String,
		value:       row.Value.String,
		description: row.Description.String,
		category:    otf.VariableCategory(row.Category.String),
		sensitive:   row.Sensitive,
		hcl:         row.HCL,
		workspaceID: row.WorkspaceID.String,
	}
}
