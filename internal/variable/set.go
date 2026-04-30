package variable

import (
	"log/slog"

	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
)

type (
	// VariableSet is a set of variables
	VariableSet struct {
		ID           resource.TfeID `db:"variable_set_id"`
		Name         string
		Description  string
		Global       bool
		Workspaces   []resource.TfeID  `db:"workspace_ids"`
		Organization organization.Name `db:"organization_name"`
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

func (s *VariableSet) updateProperties(opts UpdateVariableSetOptions) error {
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
	return nil
}
