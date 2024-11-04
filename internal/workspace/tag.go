package workspace

import (
	"errors"
	"log/slog"

	"github.com/leg100/otf/internal/resource"
)

var ErrInvalidTagSpec = errors.New("invalid tag spec: must provide either an ID or a name")

type (
	// Tag is a symbol associated with one or more workspaces. Helps searching and
	// grouping workspaces.
	Tag struct {
		ID            resource.ID // ID of the form 'tag-*'. Globally unique.
		Name          string      // Meaningful symbol. Unique to an organization.
		InstanceCount int         // Number of workspaces that have this tag
		Organization  string      // Organization this tag belongs to.
	}

	// TagSpec specifies a tag. Either ID or Name must be non-nil for it to
	// valid.
	TagSpec struct {
		ID   resource.ID
		Name string
	}

	TagSpecs []TagSpec
)

func (s TagSpec) Valid() error {
	if s.ID == resource.EmptyID && s.Name == "" {
		return ErrInvalidTagSpec
	}
	return nil
}

func (specs TagSpecs) LogValue() slog.Value {
	var (
		ids   = make([]resource.ID, len(specs))
		names = make([]string, len(specs))
	)
	for i, s := range specs {
		ids[i] = s.ID
		names[i] = s.Name
	}
	return slog.GroupValue(
		slog.Any("ids", ids),
		slog.Any("names", names),
	)
}
