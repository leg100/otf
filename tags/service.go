package tags

import (
	"context"
	"errors"

	"github.com/leg100/otf"
	"github.com/leg100/otf/logr"
)

type (
	TagService = Service

	Service interface {
		// ListTags lists tags within an organization
		ListTags(ctx context.Context, organization string, opts ListTagsOptions) (*TagList, error)

		// DeleteTags deletes tags from an organization
		DeleteTags(ctx context.Context, organization string, tagIDs []string) error

		// TagWorkspaces adds an existing tag to a list of workspaces
		TagWorkspaces(ctx context.Context, tagID string, workspaceIDs []string) error

		// AddTags appends tags to a workspace. Any tag specified by ID must
		// exist. Any tag specified by name is created if it does not
		// exist.
		AddTags(ctx context.Context, workspaceID string, tags []TagSpec) error

		// RemoveTags removes tags from a workspace. The workspace must already
		// exist. Any tag specifying an ID must exist. Any tag specifying a name
		// need not exist and no action is taken. If a tag is no longer
		// associated with any workspaces it is removed.
		RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error

		// ListWorkspaceTags lists the tags for a workspace.
		ListWorkspaceTags(ctx context.Context, workspaceID string, options ListWorkspaceTagsOptions) (*TagList, error)
	}

	service struct {
		logr.Logger
		db *pgdb
	}

	Options struct {
		otf.DB
		logr.Logger
	}

	// ListTagsOptions are options for paginating and filtering a list of
	// tags
	ListTagsOptions struct {
		otf.ListOptions
	}

	// ListWorkspaceTagsOptions are options for paginating and filtering a list of
	// workspace tags
	ListWorkspaceTagsOptions struct {
		otf.ListOptions
	}
)

func NewService(opts Options) *service {
	return &service{
		db:     &pgdb{opts.DB},
		Logger: opts.Logger,
	}
}

func (s *service) ListTags(ctx context.Context, organization string, opts ListTagsOptions) (*TagList, error) {
	list, err := s.db.listTags(ctx, organization, opts)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *service) DeleteTags(ctx context.Context, organization string, tagIDs []string) error {
	if err := s.db.deleteTags(ctx, organization, tagIDs); err != nil {
		return err
	}
	return nil
}

func (s *service) TagWorkspaces(ctx context.Context, tagID string, workspaceIDs []string) error {
	if err := s.db.tagWorkspaces(ctx, tagID, workspaceIDs); err != nil {
		return err
	}
	return nil
}

func (s *service) AddTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	// For each tag:
	// (i) if specified by name, create new tag if it does not exist and get its ID.
	// (ii) add tag to workspace
	err := s.db.tx(ctx, func(tx *pgdb) error {
		for _, t := range tags {
			var tagID string

			switch {
			case t.Name != nil:
				tagID = otf.NewID("tag")

				// creates tag if it doesn't exist.
				err := tx.addTag(ctx, workspaceID, &Tag{
					ID:   tagID,
					Name: *t.Name,
				})
				if err != nil {
					return err
				}
			case t.ID != nil:
				tagID = *t.ID
			default:
				return ErrInvalidTagSpec
			}
			if err := tx.addWorkspaceTag(ctx, workspaceID, tagID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	err := s.db.lock(ctx, func(tx *pgdb) (err error) {
		for _, t := range tags {
			var tag *Tag

			switch {
			case t.Name != nil:
				tag, err = tx.findTagByName(ctx, *t.Name)
				if errors.Is(err, otf.ErrResourceNotFound) {
					continue
				} else if err != nil {
					return err
				}
			case t.ID != nil:
				tag, err = tx.findTagByID(ctx, *t.ID)
				if err != nil {
					return err
				}
			default:
				return otf.ErrInvalidName
			}
			if err := tx.deleteWorkspaceTag(ctx, workspaceID, tag.ID); err != nil {
				return err
			}
			// Delete tag if it is no longer associated with any workspaces. If
			// that is the case then instance count should be 1, since its last
			// workspace has just been deleted.
			if tag.InstanceCount == 1 {
				if err := tx.deleteTag(ctx, tag); err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *service) ListWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*TagList, error) {
	list, err := s.db.listWorkspaceTags(ctx, workspaceID, opts)
	if err != nil {
		return nil, err
	}
	return list, nil
}
