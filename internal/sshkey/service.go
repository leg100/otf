package sshkey

import (
	"context"

	"github.com/gorilla/mux"
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
		tfe *tfe
		api *api
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
	svc.api = &api{
		Service: svc,
	}
	svc.tfe = &tfe{
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
	s.tfe.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, opts CreateOptions) (*SSHKey, error) {
	subject, err := s.Authorize(ctx, authz.CreateSSHKeyAction, opts.Organization)
	if err != nil {
		return nil, err
	}
	key, err := func() (*SSHKey, error) {
		key, privateKey, err := New(opts)
		if err != nil {
			return nil, err
		}
		return key, s.db.create(ctx, key, privateKey)
	}()
	if err != nil {
		s.Error(err, "creating ssh key", "subject", subject)
		return nil, err
	}
	s.V(0).Info("created ssh key", "key", key, "subject", subject)
	return key, nil
}

func (s *Service) GetPrivateKey(ctx context.Context, id resource.TfeID) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.GetPrivateKeySSHKeyAction, id)
	if err != nil {
		return nil, err
	}
	key, err := s.db.getPrivateKey(ctx, id)
	if err != nil {
		s.Error(err, "retrieving ssh private key", "key_id", id, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved ssh private key", "key_id", id, "subject", subject)
	return key, nil
}

func (s *Service) Get(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	subject, err := s.Authorize(ctx, authz.GetSSHKeyAction, id)
	if err != nil {
		return nil, err
	}
	key, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving ssh key", "key", key.ID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved ssh key", "key", key, "subject", subject)
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
		return nil
	})
	if err != nil {
		s.Error(err, "updating ssh key", "id", id, "subject", subject)
		return nil, err
	}
	s.V(0).Info("updated ssh key", "key", updated, "subject", subject)
	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id resource.TfeID) (*SSHKey, error) {
	subject, err := s.Authorize(ctx, authz.DeleteSSHKeyAction, id)
	if err != nil {
		return nil, err
	}
	key, err := s.db.delete(ctx, id)
	if err != nil {
		s.Error(err, "deleting ssh key", "id", id, "subject", subject)
		return nil, err
	}
	s.V(0).Info("deleted ssh key", "key", key, "subject", subject)
	return key, nil
}
