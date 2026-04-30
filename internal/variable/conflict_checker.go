package variable

import (
	"context"

	"github.com/leg100/otf/internal/resource"
)

// conflictChecker checks whether a variable conflicts with another variable. A
// conflict occurs if two variables within a workspace or a set share the same
// key and category, or if a variable within a global set shares the same key
// and category as another global set.
type conflictChecker struct {
	db *pgdb
}

type conflictCheckerClient interface {
}

// checkVariable checks whether a variable conflicts with other variables in the
// same scope. If the variable belongs to a workspace then the scope is other
// variables that belong to the workspace. If the variable belongs to a set then
// the scope is other variables that belong to the set, and if it's a global set
// then scope is expanded to variables in other global sets in the same
// organization.
func (c *conflictChecker) checkVariable(ctx context.Context, v *Variable) (err error) {
	var scopedVariables []*Variable

	var set *VariableSet
	if v.ParentID.Kind() == resource.VariableSetKind {
		set, err = c.db.getVariableSet(ctx, v.ParentID)
		if err != nil {
			return err
		}
	}
	if set != nil && set.Global {
		scopedVariables, err = c.db.listGlobalVariables(ctx, set.Organization)
		if err != nil {
			return err
		}
	} else {
		scopedVariables, err = c.db.listVariables(ctx, v.ParentID)
		if err != nil {
			return err
		}
	}
	if err := v.checkConflicts(scopedVariables); err != nil {
		return err
	}
	return nil
}

// checkSet checks whether the variable set's variables conflicts with a
// variable in any other set. Note that they only do so if the set is global and
// the other variable is in a global set.
func (c *conflictChecker) checkSet(ctx context.Context, set *VariableSet) error {
	if !set.Global {
		return nil
	}
	setVariables, err := c.db.listVariables(ctx, set.ID)
	if err != nil {
		return err
	}
	scopedVariables, err := c.db.listGlobalVariables(ctx, set.Organization)
	if err != nil {
		return err
	}
	for _, v1 := range setVariables {
		if err := v1.checkConflicts(scopedVariables); err != nil {
			return err
		}
	}
	return nil
}
