package workspace

import (
	"context"

	"github.com/leg100/otf/internal/rbac"
)

type (
	RemoteStateConsumerService interface {
		// ListRemoteStateConsumers lists remote state consumers for a
		// workspace.
		ListRemoteStateConsumers(ctx context.Context, workspaceID string) ([]*Workspace, error)

		// RemoteStateConsumerWorkspaces replaces a workspace's list of remote
		// state consumers. The workspace ID should be given for each consumer.
		ReplaceRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error

		// AddRemoteStateConsumers adds one or more remote state consumers to a
		// workspace.
		AddRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error

		// DeleteRemoteStateConsumers deletes one or more remote state consumers from a
		// workspace.
		DeleteRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error
	}
)

func (s *service) ListRemoteStateConsumers(ctx context.Context, workspaceID string) ([]*Workspace, error) {
	subject, err := s.organization.CanAccess(ctx, rbac.ListRemoteStateConsumersAction, workspaceID)
	if err != nil {
		return nil, err
	}

	list, err := s.db.listRemoteStateConsumers(ctx, workspaceID)
	if err != nil {
		s.Error(err, "listing remote state consumers", "workspace_id", workspaceID, "subject", subject)
	}
	s.V(9).Info("listed remote state consumers", "workspace_id", workspaceID, "subject", subject)
	return list, nil
}

func (s *service) ReplaceRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	subject, err := s.organization.CanAccess(ctx, rbac.ReplaceRemoteStateConsumersAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.replaceRemoteStateConsumers(ctx, workspaceID, consumers); err != nil {
		return err
	}
	if err != nil {
		s.Error(err, "tagging remote state consumers", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
		return err
	}
	s.Info("tagged workspaces", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
	return nil
}

func (s *service) AddRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	subject, err := s.organization.CanAccess(ctx, rbac.AddRemoteStateConsumersAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.addRemoteStateConsumers(ctx, workspaceID, consumers); err != nil {
		s.Error(err, "adding remote state consumers", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
		return err
	}
	s.Info("added remote state consumers", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
	return nil
}

func (s *service) DeleteRemoteStateConsumers(ctx context.Context, workspaceID string, consumers []string) error {
	subject, err := s.organization.CanAccess(ctx, rbac.DeleteRemoteStateConsumersAction, workspaceID)
	if err != nil {
		return err
	}

	if err := s.db.deleteRemoteStateConsumers(ctx, workspaceID, consumers); err != nil {
		s.Error(err, "deleting remote state consumers", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
		return err
	}
	s.Info("deleted remote state consumers", "workspace_id", workspaceID, "consumers", consumers, "subject", subject)
	return nil
}
