package variable

import (
	"fmt"
	"log/slog"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type (
	// VariableSet is a set of variables
	VariableSet struct {
		ID           resource.TfeID
		Name         string
		Description  string
		Global       bool
		Workspaces   []resource.TfeID
		Organization organization.Name
		Variables    []*Variable
	}

	CreateVariableSetOptions struct {
		Name        string
		Description string
		Global      bool
		Workspaces  []resource.TfeID
	}

	UpdateVariableSetOptions struct {
		Name        *string
		Description *string
		Global      *bool
		Workspaces  []resource.TfeID
	}

	ListOptions struct {
		resource.PageOptions
		Organization organization.Name `schema:"organization_name"`
	}
)

func newSet(organization organization.Name, opts CreateVariableSetOptions) (*VariableSet, error) {
	return &VariableSet{
		ID:           resource.NewTfeID(resource.VariableSetKind),
		Name:         opts.Name,
		Description:  opts.Description,
		Global:       opts.Global,
		Organization: organization,
	}, nil
}

func (s *VariableSet) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", s.ID.String()),
		slog.String("name", s.Name),
		slog.Any("organization", s.Organization),
		slog.Bool("global", s.Global),
		slog.Any("workspaces", s.Workspaces),
	}
	return slog.GroupValue(attrs...)
}

func (s *VariableSet) addVariable(organizationSets []*VariableSet, opts CreateVariableOptions) (*Variable, error) {
	v, err := newVariable(s.Variables, opts)
	if err != nil {
		return nil, err
	}
	if err := s.checkGlobalConflicts(organizationSets); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VariableSet) updateVariable(organizationSets []*VariableSet, variableID resource.TfeID, opts UpdateVariableOptions) (*Variable, error) {
	v := s.GetVariableByID(variableID)
	if v == nil {
		return nil, fmt.Errorf("cannot find variable %s in set", v.ID)
	}
	if err := v.update(s.Variables, opts); err != nil {
		return nil, err
	}
	if err := s.checkGlobalConflicts(organizationSets); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VariableSet) updateProperties(organizationSets []*VariableSet, opts UpdateVariableSetOptions) error {
	if opts.Name != nil {
		s.Name = *opts.Name
	}
	if opts.Description != nil {
		s.Description = *opts.Description
	}
	if opts.Global != nil {
		s.Global = *opts.Global
	}
	if opts.Workspaces != nil {
		s.Workspaces = opts.Workspaces
	}
	if err := s.checkGlobalConflicts(organizationSets); err != nil {
		return err
	}
	return nil
}

func (s *VariableSet) GetVariableByID(variableID resource.TfeID) *Variable {
	for _, v := range s.Variables {
		if v.ID == variableID {
			return v
		}
	}
	return nil
}

// checkGlobalConflicts checks for variable conflicts within not only the set,
// but with the other given sets too. If any of the following is true, then
// ErrVariableConflict is returned:
//
// (a) set contains more than one variable sharing the same key and category
// (b) set is global and contains a variable that shares the same key and category as another
// variable in another global set in the given sets
func (s *VariableSet) checkGlobalConflicts(organizationSets []*VariableSet) error {
	if !s.Global {
		// only global sets conflict with one another
		return nil
	}
	for _, other := range organizationSets {
		if s.ID == other.ID {
			// skip same variable set
			continue
		}
		if !other.Global {
			// set can only conflict with other global sets
			continue
		}
		// check for conflicts between each set variable and each variable in all
		// the global sets
		for _, v := range s.Variables {
			if err := v.checkConflicts(other.Variables); err != nil {
				return err
			}
		}
	}
	return nil
}
