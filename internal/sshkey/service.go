package sshkey

import (
	"context"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/organization"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db  *pgdb
		api *tfe
	}

	Options struct {
		*sql.DB
		*tfeapi.Responder
		logr.Logger

		Authorizer *authz.Authorizer
	}
)

func NewService(opts Options) *Service {
	svc := &Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
		db:         &pgdb{opts.DB},
	}
	svc.api = &tfe{
		Service:   svc,
		Responder: opts.Responder,
	}
	// Register parent resolver so the authorizer can resolve ssh key -> org
	opts.Authorizer.RegisterParentResolver(resource.SSHKeyKind,
		func(ctx context.Context, id resource.ID) (resource.ID, error) {
			key, err := svc.db.get(ctx, id.(resource.TfeID))
			if err != nil {
				return nil, err
			}
			return key.Organization, nil
		},
	)
	return svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, opts CreateOptions) (*SSHKey, error) {
	if opts.Organization == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "organization"}
	}
	if opts.Name == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "name"}
	}
	if opts.PrivateKey == nil {
		return nil, &internal.ErrMissingParameter{Parameter: "value"}
	}
	subject, err := s.Authorize(ctx, authz.CreateSSHKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}
	key := &SSHKey{
		ID:           resource.NewTfeID(resource.SSHKeyKind),
		CreatedAt:    internal.CurrentTimestamp(nil),
		UpdatedAt:    internal.CurrentTimestamp(nil),
		Name:         *opts.Name,
		Organization: *opts.Organization,
		PrivateKey:   *opts.PrivateKey,
	}
	if err := s.db.create(ctx, key); err != nil {
		s.Error(err, "creating ssh key", "subject", subject)
		return nil, err
	}
	s.V(0).Info("created ssh key", "key", key.ID, "subject", subject)
	return key, nil
}

// GetSSHKey is an alias for Get, satisfying the runner.sshKeyClient interface.
func (s *Service) GetSSHKey(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	return s.Get(ctx, id)
}

func (s *Service) Get(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	key, err := s.db.get(ctx, id)
	if err != nil {
		return nil, err
	}
	subject, err := s.Authorize(ctx, authz.GetSSHKeyAction, id)
	if err != nil {
		return nil, err
	}
	s.V(9).Info("retrieved ssh key", "key", key.ID, "subject", subject)
	return key, nil
}

func (s *Service) List(ctx context.Context, org organization.Name) ([]*SSHKey, error) {
	subject, err := s.Authorize(ctx, authz.ListSSHKeysAction, org)
	if err != nil {
		return nil, err
	}
	keys, err := s.db.list(ctx, org)
	if err != nil {
		s.Error(err, "listing ssh keys", "subject", subject)
		return nil, err
	}
	s.V(9).Info("listed ssh keys", "total", len(keys), "subject", subject)
	return keys, nil
}

func (s *Service) Update(ctx context.Context, id resource.TfeID, opts UpdateOptions) (*SSHKey, error) {
	var subject authz.Subject
	updated, err := s.db.update(ctx, id, func(ctx context.Context, key *SSHKey) (err error) {
		subject, err = s.Authorize(ctx, authz.UpdateSSHKeyAction, id)
		if err != nil {
			return err
		}
		if opts.Name != nil {
			key.Name = *opts.Name
		}
		if opts.PrivateKey != nil {
			key.PrivateKey = *opts.PrivateKey
		}
		key.UpdatedAt = internal.CurrentTimestamp(nil)
		return nil
	})
	if err != nil {
		s.Error(err, "updating ssh key", "id", id, "subject", subject)
		return nil, err
	}
	s.V(0).Info("updated ssh key", "key", updated.ID, "subject", subject)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id resource.TfeID) error {
	key, err := s.db.get(ctx, id)
	if err != nil {
		return err
	}
	subject, err := s.Authorize(ctx, authz.DeleteSSHKeyAction, id)
	if err != nil {
		return err
	}
	if err := s.db.delete(ctx, id); err != nil {
		s.Error(err, "deleting ssh key", "id", id, "subject", subject)
		return err
	}
	s.V(0).Info("deleted ssh key", "key", key.ID, "subject", subject)
	return nil
}
