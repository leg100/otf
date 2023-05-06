package workspace

import (
	"errors"

	"github.com/leg100/otf/internal"
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
		*internal.Pagination
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
	if s.ID == "" && s.Name == "" {
		return ErrInvalidTagSpec
	}
	return nil
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
