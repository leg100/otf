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
		ID            resource.ID               // ID of the form 'tag-*'. Globally unique.
		Name          string                    // Meaningful symbol. Unique to an organization.
		InstanceCount int                       // Number of workspaces that have this tag
		Organization  resource.OrganizationName // Organization this tag belongs to.
	}

	// TagSpec specifies a tag. Either ID or Name must be non-nil for it to
	// be valid.
	TagSpec struct {
		ID   resource.ID
		Name string
	}

	TagSpecs []TagSpec
)

func (s TagSpec) Valid() error {
	if s.ID == nil && s.Name == "" {
		return ErrInvalidTagSpec
	}
	return nil
}

func (specs TagSpecs) LogValue() slog.Value {
	var (
		ids   []resource.ID
		names []string
	)
	for _, s := range specs {
		if s.ID != nil {
			ids = append(ids, *s.ID)
		}
		if s.Name != "" {
			names = append(names, s.Name)
		}
	}
	return slog.GroupValue(
		slog.Any("ids", ids),
		slog.Any("names", names),
	)
}
