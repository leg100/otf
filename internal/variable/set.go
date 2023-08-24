package variable

import (
	"log/slog"
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
	}
	UpdateVariableSetOptions struct {
		Name        *string
		Description *string
		Global      *bool
	}
)

func (s *VariableSet) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.String("id", s.ID),
		slog.String("name", s.Name),
		slog.String("organization", s.Organization),
	}
	return slog.GroupValue(attrs...)
}

func (s *VariableSet) update(opts UpdateVariableSetOptions) error {
	if opts.Name != nil {
		s.Name = *opts.Name
	}
	if opts.Description != nil {
		s.Description = *opts.Description
	}
	if opts.Global != nil {
		s.Global = *opts.Global
	}
	return nil
}

func (s *VariableSet) hasVariable(variableID string) (bool, *Variable) {
	for _, v := range s.Variables {
		if v.ID == variableID {
			return true, v
		}
	}
	return false, nil
}

// conflicts determines whether the set has a variable that conflicts with v,
// i.e. has same key and category
func (s *VariableSet) conflicts(v *Variable) error {
	for _, v2 := range s.Variables {
		if err := v.conflicts(v2); err != nil {
			return err
		}
	}
	return nil
}

// checkConflicts checks whether v conflicts with another variable: s is the set
// it belongs to, or is going to belong to, and sets are the other sets in the
// same organization (which may include s). If any of the following is true,
// then ErrVariableConflict is returned:
//
// (a) v shares the same key and category with another v in s
// (b) v shares the same key and category with another v in sets, iff both s and
// the set containing the other v are both global.
func checkConflicts(v *Variable, s *VariableSet, sets []*VariableSet) error {
	if err := s.conflicts(v); err != nil {
		return err
	}
	if !s.Global {
		return nil
	}
	// check for conflicts with other global sets
	for _, s2 := range sets {
		if s2.ID == s.ID {
			// skip same variable set
			continue
		}
		if !s.Global {
			continue
		}
		if err := s.conflicts(v); err != nil {
			return err
		}
	}
	return nil
}
