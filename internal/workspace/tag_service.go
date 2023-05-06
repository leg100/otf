package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/rbac"
)

type (
	TagService interface {
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

		listAllTags(ctx context.Context, organization string) ([]*Tag, error)
	}

	// ListTagsOptions are options for paginating and filtering a list of
	// tags
	ListTagsOptions struct {
		internal.ListOptions
	}

	// ListWorkspaceTagsOptions are options for paginating and filtering a list of
	// workspace tags
	ListWorkspaceTagsOptions struct {
		internal.ListOptions
	}
)

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
	subject, err := internal.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.db.tx(ctx, func(tx *pgdb) error {
		for _, wid := range workspaceIDs {
			_, err := s.CanAccess(ctx, rbac.TagWorkspacesAction, wid)
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
	subject, err := s.CanAccess(ctx, rbac.AddTagsAction, workspaceID)
	if err != nil {
		return err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found; %s; %w", workspaceID, err)
	}

	added, err := addTags(ctx, s.db, ws, tags)
	if err != nil {
		s.Error(err, "adding tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
		return err
	}
	s.Info("added tags", "workspace", workspaceID, "tags", added, "subject", subject)
	return nil
}

func (s *service) RemoveTags(ctx context.Context, workspaceID string, tags []TagSpec) error {
	subject, err := s.CanAccess(ctx, rbac.RemoveTagsAction, workspaceID)
	if err != nil {
		return err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found; %s; %w", workspaceID, err)
	}

	err = s.db.lockTags(ctx, func(tx *pgdb) (err error) {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return err
			}
			var tag *Tag

			switch {
			case t.Name != "":
				tag, err = tx.findTagByName(ctx, ws.Organization, t.Name)
				if errors.Is(err, internal.ErrResourceNotFound) {
					// ignore tags that cannot be found when specified by name
					continue
				} else if err != nil {
					return err
				}
			case t.ID != "":
				tag, err = tx.findTagByID(ctx, ws.Organization, t.ID)
				if err != nil {
					return err
				}
			default:
				return ErrInvalidTagSpec
			}
			if err := tx.deleteWorkspaceTag(ctx, workspaceID, tag.ID); err != nil {
				return fmt.Errorf("removing tag %s from workspace %s: %w", tag.ID, workspaceID, err)
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
	subject, err := s.CanAccess(ctx, rbac.ListWorkspaceTags, workspaceID)
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

func (s *service) listAllTags(ctx context.Context, organization string) ([]*Tag, error) {
	var (
		tags []*Tag
		opts ListTagsOptions
	)
	for {
		page, err := s.ListTags(ctx, organization, opts)
		if err != nil {
			return nil, err
		}
		tags = append(tags, page.Items...)
		if page.NextPage() == nil {
			break
		}
		opts.PageNumber = *page.NextPage()
	}
	return tags, nil
}

func addTags(ctx context.Context, db *pgdb, ws *Workspace, tags []TagSpec) ([]string, error) {
	// For each tag:
	// (i) if specified by name, create new tag if it does not exist and get its ID.
	// (ii) add tag to workspace
	var added []string
	err := db.lockTags(ctx, func(tx *pgdb) (err error) {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return err
			}

			id := t.ID
			name := t.Name
			switch {
			case name != "":
				existing, err := tx.findTagByName(ctx, ws.Organization, name)
				if errors.Is(err, internal.ErrResourceNotFound) {
					id = internal.NewID("tag")
					if err := tx.addTag(ctx, ws.Organization, name, id); err != nil {
						return err
					}
				} else if err != nil {
					return err
				} else {
					id = existing.ID
				}
			case id != "":
				existing, err := tx.findTagByID(ctx, ws.Organization, t.ID)
				if err != nil {
					return err
				}
				name = existing.Name
			default:
				return ErrInvalidTagSpec
			}

			if err := tx.tagWorkspace(ctx, ws.ID, id); err != nil {
				return err
			}
			added = append(added, name)
		}
		return nil
	})
	return added, err
}
