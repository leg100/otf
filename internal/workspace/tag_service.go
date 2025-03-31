package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
)

type (
	// ListTagsOptions are options for paginating and filtering a list of
	// tags
	ListTagsOptions struct {
		resource.PageOptions
	}

	// ListWorkspaceTagsOptions are options for paginating and filtering a list of
	// workspace tags
	ListWorkspaceTagsOptions struct {
		resource.PageOptions
	}
)

func (s *Service) ListTags(ctx context.Context, organization resource.OrganizationName, opts ListTagsOptions) (*resource.Page[*Tag], error) {
	subject, err := s.Authorize(ctx, authz.ListTagsAction, organization)
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

func (s *Service) DeleteTags(ctx context.Context, organization resource.OrganizationName, tagIDs []resource.TfeID) error {
	subject, err := s.Authorize(ctx, authz.DeleteTagsAction, organization)
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

func (s *Service) TagWorkspaces(ctx context.Context, tagID resource.TfeID, workspaceIDs []resource.TfeID) error {
	subject, err := authz.SubjectFromContext(ctx)
	if err != nil {
		return err
	}

	err = s.db.Tx(ctx, func(ctx context.Context, _ sql.Connection) error {
		for _, wid := range workspaceIDs {
			_, err := s.Authorize(ctx, authz.TagWorkspacesAction, wid)
			if err != nil {
				return err
			}
			if err := s.db.tagWorkspace(ctx, wid, tagID); err != nil {
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

func (s *Service) AddTags(ctx context.Context, workspaceID resource.TfeID, tags []TagSpec) error {
	subject, err := s.Authorize(ctx, authz.AddTagsAction, workspaceID)
	if err != nil {
		return err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found; %s; %w", workspaceID, err)
	}

	added, err := s.addTags(ctx, ws, tags)
	if err != nil {
		s.Error(err, "adding tags", "workspace", workspaceID, "tags", TagSpecs(tags), "subject", subject)
		return err
	}
	s.Info("added tags", "workspace", workspaceID, "tags", added, "subject", subject)
	return nil
}

func (s *Service) RemoveTags(ctx context.Context, workspaceID resource.TfeID, tags []TagSpec) error {
	subject, err := s.Authorize(ctx, authz.RemoveTagsAction, workspaceID)
	if err != nil {
		return err
	}

	ws, err := s.db.get(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("workspace not found; %s; %w", workspaceID, err)
	}

	err = s.db.Lock(ctx, "tags", func(ctx context.Context, _ sql.Connection) (err error) {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return err
			}
			var tag *Tag

			switch {
			case t.Name != "":
				tag, err = s.db.findTagByName(ctx, ws.Organization, t.Name)
				if errors.Is(err, internal.ErrResourceNotFound) {
					// ignore tags that cannot be found when specified by name
					continue
				} else if err != nil {
					return err
				}
			case t.ID != nil:
				tag, err = s.db.findTagByID(ctx, ws.Organization, *t.ID)
				if err != nil {
					return err
				}
			default:
				return ErrInvalidTagSpec
			}
			if err := s.db.deleteWorkspaceTag(ctx, workspaceID, tag.ID); err != nil {
				return fmt.Errorf("removing tag %s from workspace %s: %w", tag.ID, workspaceID, err)
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

func (s *Service) ListWorkspaceTags(ctx context.Context, workspaceID resource.TfeID, opts ListWorkspaceTagsOptions) (*resource.Page[*Tag], error) {
	subject, err := s.Authorize(ctx, authz.ListWorkspaceTags, &authz.Request{ID: &workspaceID})
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

func (s *Service) addTags(ctx context.Context, ws *Workspace, tags []TagSpec) ([]string, error) {
	// For each tag:
	// (i) if specified by name, create new tag if it does not exist and get its ID.
	// (ii) add tag to workspace
	var added []string
	err := s.db.Lock(ctx, "tags", func(ctx context.Context, _ sql.Connection) (err error) {
		for _, t := range tags {
			if err := t.Valid(); err != nil {
				return fmt.Errorf("invalid tag: %w", err)
			}

			id := t.ID
			name := t.Name
			switch {
			case name != "":
				existing, err := s.db.findTagByName(ctx, ws.Organization, name)
				if errors.Is(err, internal.ErrResourceNotFound) {
					idValue := resource.NewTfeID("tag")
					id = &idValue
					if err := s.db.addTag(ctx, ws.Organization, name, *id); err != nil {
						return fmt.Errorf("adding tag: %s %w", name, err)
					}
				} else if err != nil {
					return err
				} else {
					id = &existing.ID
				}
			case id != nil:
				existing, err := s.db.findTagByID(ctx, ws.Organization, *t.ID)
				if err != nil {
					return err
				}
				name = existing.Name
			default:
				return ErrInvalidTagSpec
			}

			if err := s.db.tagWorkspace(ctx, ws.ID, *id); err != nil {
				return err
			}
			added = append(added, name)
		}
		return nil
	})
	return added, err
}
