package variable

import (
	"fmt"
	"log/slog"

	"github.com/leg100/otf/internal"
)

type (
	// VariableSet is a set of variables
	VariableSet struct {
		ID           string
		Name         string
		Description  string
		Global       bool
		Variables    []*Variable
		Workspaces   []string // workspace IDs
		Organization string   // org name
	}
	CreateVariableSetOptions struct {
		Name        string
		Description string
		Global      bool
		Workspaces  []string // workspace IDs
	}
	UpdateVariableSetOptions struct {
		Name        *string
		Description *string
		Global      *bool
		Workspaces  []string // workspace IDs
	}
)

func (s *VariableSet) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", s.ID),
		slog.String("name", s.Name),
		slog.String("organization", s.Organization),
		slog.Bool("global", s.Global),
		slog.Any("workspaces", s.Workspaces),
	}
	return slog.GroupValue(attrs...)
}

func newSet(organization string, opts CreateVariableSetOptions) (*VariableSet, error) {
	return &VariableSet{
		ID:           internal.NewID("varset"),
		Name:         opts.Name,
		Description:  opts.Description,
		Global:       opts.Global,
		Organization: organization,
	}, nil
}

func (s *VariableSet) update(sets []*VariableSet, opts UpdateVariableSetOptions) error {
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
	if err := s.checkConflicts(sets); err != nil {
		return err
	}
	return nil
}

func (s *VariableSet) createVariable(sets []*VariableSet, opts CreateVariableOptions) (*Variable, error) {
	v, err := newVariable(opts)
	if err != nil {
		return nil, err
	}
	s.Variables = append(s.Variables, v)
	if err := s.checkConflicts(sets); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VariableSet) updateVariable(variableID string, sets []*VariableSet, opts UpdateVariableOptions) (*Variable, error) {
	v := s.getVariable(variableID)
	if v == nil {
		return nil, fmt.Errorf("cannot find variable %s in set", v.ID)
	}
	if err := v.update(s.Variables, opts); err != nil {
		return nil, err
	}
	if err := s.checkConflicts(sets); err != nil {
		return nil, err
	}
	return v, nil
}

func (s *VariableSet) getVariable(variableID string) *Variable {
	for _, v := range s.Variables {
		if v.ID == variableID {
			return v
		}
	}
	return nil
}

// checkConflicts checks for variable conflicts within not only the set, but
// with the other given sets too. If any of the following is true, then
// ErrVariableConflict is returned:
//
// (a) set contains more than one variable sharing the same key and category
// (b) set is global and contains a variable that shares the same key and category as another
// variable in another global set in the given sets
func (s *VariableSet) checkConflicts(sets []*VariableSet) error {
	// check no variable in set conflicts with another variable in set
	for _, v := range s.Variables {
		if err := v.conflicts(s.Variables); err != nil {
			return err
		}
	}
	if !s.Global {
		return nil
	}
	// check conflicts with other global sets
	for _, other := range sets {
		if !s.Global {
			continue
		}
		if other.ID == s.ID {
			// skip same variable set
			continue
		}
		for _, v := range s.Variables {
			if err := v.conflicts(other.Variables); err != nil {
				return err
			}
		}
	}
	return nil
}
