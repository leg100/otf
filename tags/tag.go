// Package tags handles tagging of workspaces.
package tags

import (
	"errors"

	"github.com/leg100/otf"
	"golang.org/x/exp/slog"
)

var ErrInvalidTagSpec = errors.New("invalid tag spec: must provide either an ID or a name")

type (
	// Tag is a symbol associated with one or more workspaces. Helps searching and
	// grouping workspaces.
	Tag struct {
		ID            string // ID of the form 'tag-*'. Globally unique.
		Name          string // Meaningful symbol. Unique to an organization.
		InstanceCount int    // Number of workspaces that have this tag
		Organization  string // Organization this tag belongs to.
	}

	// TagList is a list of tags.
	TagList struct {
		*otf.Pagination
		Items []*Tag
	}

	// TagSpec specifies a tag. Either ID or Name must be non-nil for it to
	// valid.
	TagSpec struct {
		ID   string
		Name string
	}

	TagSpecs []TagSpec
)

func (s TagSpec) Valid() error {
	if s.ID != "" {
		return nil
	} else if s.Name != "" {
		return nil
	} else {
		return ErrInvalidTagSpec
	}
}

func (specs TagSpecs) LogValue() slog.Value {
	var (
		ids   []string
		names []string
	)
	for _, s := range specs {
		ids = append(ids, s.ID)
		names = append(names, s.Name)
	}
	return slog.GroupValue(
		slog.Any("ids", ids),
		slog.Any("names", names),
	)
}
