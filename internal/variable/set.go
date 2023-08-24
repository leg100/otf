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

func (s *VariableSet) hasVariable(variableID string) bool {
	for _, v := range s.Variables {
		if v.ID == variableID {
			return true
		}
	}
	return false
}
