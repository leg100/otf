package variable

const (
	CategoryTerraform VariableCategory = "terraform"
	CategoryEnv       VariableCategory = "env"
)

// VariableCategory is the category of variable
type VariableCategory string

// VariableCategoryPtr returns a pointer to the given category type.
func VariableCategoryPtr(v VariableCategory) *VariableCategory {
	return &v
}
