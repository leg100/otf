package sql

import (
	"context"

	"github.com/leg100/otf"
	"github.com/leg100/otf/sql/pggen"
)

func (db *DB) CreateVariable(ctx context.Context, ws *otf.Variable) error {
	_, err := db.InsertVariable(ctx, pggen.InsertVariableParams{
		VariableID:  String(ws.ID()),
		Key:         String(ws.Key()),
		Value:       String(ws.Value()),
		Description: String(ws.Description()),
		Category:    String(string(ws.Category())),
		Sensitive:   ws.Sensitive(),
		HCL:         ws.HCL(),
		WorkspaceID: String(ws.WorkspaceID()),
	})
	if err != nil {
		return Error(err)
	}
	return nil
}

func (db *DB) UpdateVariable(ctx context.Context, variableID string, fn func(*otf.Variable) error) (*otf.Variable, error) {
	var variable *otf.Variable
	err := db.tx(ctx, func(tx *DB) error {
		var err error
		// retrieve variable
		row, err := tx.FindVariableForUpdate(ctx, String(variableID))
		if err != nil {
			return Error(err)
		}
		variable = otf.UnmarshalVariableRow(otf.VariableRow(row))

		// update variable
		if err := fn(variable); err != nil {
			return err
		}
		// persist variable
		_, err = tx.Querier.UpdateVariableByID(ctx, pggen.UpdateVariableByIDParams{
			VariableID:  String(variableID),
			Key:         String(variable.Key()),
			Value:       String(variable.Value()),
			Description: String(variable.Description()),
			Category:    String(string(variable.Category())),
			Sensitive:   variable.Sensitive(),
			HCL:         variable.HCL(),
		})
		return err
	})
	return variable, err
}

func (db *DB) ListVariables(ctx context.Context, workspaceID string) ([]*otf.Variable, error) {
	rows, err := db.FindVariables(ctx, String(workspaceID))
	if err != nil {
		return nil, err
	}

	var variables []*otf.Variable
	for _, row := range rows {
		variables = append(variables, otf.UnmarshalVariableRow(otf.VariableRow(row)))
	}
	return variables, nil
}

func (db *DB) GetVariable(ctx context.Context, variableID string) (*otf.Variable, error) {
	row, err := db.FindVariable(ctx, String(variableID))
	if err != nil {
		return nil, err
	}

	return otf.UnmarshalVariableRow(otf.VariableRow(row)), nil
}

func (db *DB) DeleteVariable(ctx context.Context, variableID string) (*otf.Variable, error) {
	row, err := db.Querier.DeleteVariableByID(ctx, String(variableID))
	if err != nil {
		return nil, Error(err)
	}
	return otf.UnmarshalVariableRow(otf.VariableRow(row)), nil
}
