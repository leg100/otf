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
	// AddHandlers adds handlers for the http api.
	AddHandlers(*mux.Router)
	// AddHTMLHandlers adds handlers for the web ui.
	AddHTMLHandlers(*mux.Router)

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
