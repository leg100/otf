package otf

import (
	"context"

	"github.com/gorilla/mux"
)

// VariableCategory is the category of variable
type VariableCategory string

const (
	CategoryTerraform VariableCategory = "terraform"
	CategoryEnv       VariableCategory = "env"
)

type VariableService interface {
	// AddHandlers adds http handlers for to the given mux. The handlers
	// implement the variable service API.
	AddHandlers(*mux.Router)

	VariableApp
}

type VariableApp interface {
	ListVariables(ctx context.Context, workspaceID string) ([]Variable, error)
}

type (
	Variable interface {
		ID() string
		Key() string
		Value() string
		Description() string
		Category() VariableCategory
		Sensitive() bool
		HCL() bool
		WorkspaceID() string
	}
	CreateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
	UpdateVariableOptions struct {
		Key         *string
		Value       *string
		Description *string
		Category    *VariableCategory
		Sensitive   *bool
		HCL         *bool
	}
)
