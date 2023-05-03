package tags

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf"
	"github.com/leg100/otf/logr"
	"github.com/leg100/otf/organization"
	"github.com/leg100/otf/rbac"
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

		workspace    otf.Authorizer // workspace authorizer
		organization otf.Authorizer // organization authorizer
	}

	Options struct {
		otf.DB
		logr.Logger

		WorkspaceAuthorizer otf.Authorizer // workspace authorizer
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
		db:           &pgdb{opts.DB},
		organization: &organization.Authorizer{Logger: opts.Logger},
		Logger:       opts.Logger,
	}
}

func (s *service) ListTags(ctx context.Context, organization string, opts ListTagsOptions) (*TagList, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.ListTagsAction, organization)
	if err != nil {
		return nil, err
	}

	list, err := s.db.listTags(ctx, organization, opts)
	if err != nil {
		s.Error(err, "listing tags", "organization", organization, "subject", subject)
	}
	s.V(9).Info("listed tags", "organization", organization, "subject", subject)
	return list, nil
}

func (s *service) DeleteTags(ctx context.Context, organization string, tagIDs []string) error {
	subject, err := s.organization.CanAccess(ctx, rbac.DeleteTagsAction, organization)
	if err != nil {
		return err
	}

	if err := s.db.deleteTags(ctx, organization, tagIDs); err != nil {
		s.Error(err, "deleting tags", "organization", organization, "tags_ids", tagIDs, "subject", subject)
		return err
	}
	s.Info("deleted tags", "organization", organization, "tag_ids", tagIDs, "subject", subject)
	return nil
}

func (s *service) TagWorkspaces(ctx context.Context, tagID string, workspaceIDs []string) error {
	subject, err := otf.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.db.tx(ctx, func(tx *pgdb) error {
		for _, wid := range workspaceIDs {
			_, err := s.workspace.CanAccess(ctx, rbac.TagWorkspacesAction, wid)
			if err != nil {
				return err
			}
			if err := tx.tagWorkspace(ctx, wid, tagID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "tagging tags", "tag_id", tagID, "workspace_ids", workspaceIDs, "subject", subject)
		return err
	}
	s.Info("tagged workspaces", "tag_id", tagID, "workspaces_ids", workspaceIDs, "subject", subject)
	return nil
}

func (s *service) AddTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	subject, err := s.workspace.CanAccess(ctx, rbac.AddTagsAction, workspaceID)
	if err != nil {
		return err
	}

	// For each tag:
	// (i) if specified by name, create new tag if it does not exist and get its ID.
	// (ii) add tag to workspace
	err = s.db.tx(ctx, func(tx *pgdb) error {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return err
			}
			var tagID string

			switch {
			case t.Name != "":
				tagID = otf.NewID("tag")

				// creates tag if it doesn't exist.
				err := tx.addTag(ctx, workspaceID, &Tag{
					ID:   tagID,
					Name: t.Name,
				})
				if err != nil {
					return err
				}
			case t.ID != "":
				tagID = t.ID
			default:
				return ErrInvalidTagSpec
			}
			if err := tx.tagWorkspace(ctx, workspaceID, tagID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "adding tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
		return err
	}
	s.Info("added tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
	return nil
}

func (s *service) RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	subject, err := s.workspace.CanAccess(ctx, rbac.RemoveTagsAction, workspaceID)
	if err != nil {
		return err
	}

	err = s.db.lock(ctx, func(tx *pgdb) (err error) {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return err
			}
			var tag *Tag

			switch {
			case t.Name != "":
				tag, err = tx.findTagByName(ctx, workspaceID, t.Name)
				if errors.Is(err, otf.ErrResourceNotFound) {
					continue
				} else if err != nil {
					return err
				}
			case t.ID != "":
				tag, err = tx.findTagByID(ctx, workspaceID, t.ID)
				if err != nil {
					return err
				}
			default:
				return ErrInvalidTagSpec
			}
			if err := tx.deleteWorkspaceTag(ctx, workspaceID, tag.ID); err != nil {
				return fmt.Errorf("deleting tag %s from workspace %s: %w", tag.ID, workspaceID, err)
			}
			// Delete tag if it is no longer associated with any workspaces. If
			// that is the case then instance count should be 1, since its last
			// workspace has just been deleted.
			if tag.InstanceCount == 1 {
				if err := tx.deleteTag(ctx, tag); err != nil {
					return fmt.Errorf("deleting tag: %w", err)
				}
			}
		}
		return nil
	})
	if err != nil {
		s.Error(err, "removing tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
		return err
	}
	s.Info("removed tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
	return nil
}

func (s *service) ListWorkspaceTags(ctx context.Context, workspaceID string, opts ListWorkspaceTagsOptions) (*TagList, error) {
	subject, err := s.workspace.CanAccess(ctx, rbac.ListWorkspaceTags, workspaceID)
	if err != nil {
		return nil, err
	}

	list, err := s.db.listWorkspaceTags(ctx, workspaceID, opts)
	if err != nil {
		s.Error(err, "listing workspace tags", "workspace", workspaceID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed workspace tags", "workspace", workspaceID, "subject", subject)
	return list, nil
}
