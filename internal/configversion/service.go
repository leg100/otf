package configversion

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/surl/v2"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer

		db     *pgdb
		cache  internal.Cache
		tfeapi *tfe
		api    *api
	}

	Options struct {
		logr.Logger

		MaxConfigSize int64
		Authorizer    *authz.Authorizer

		internal.Cache
		*sql.DB
		*surl.Signer
		*tfeapi.Responder
	}
)

func NewService(opts Options) *Service {
	svc := Service{
		Logger:     opts.Logger,
		Authorizer: opts.Authorizer,
	}
	svc.db = &pgdb{opts.DB}
	svc.cache = opts.Cache
	svc.tfeapi = &tfe{
		Logger:        opts.Logger,
		tfeClient:     &svc,
		Signer:        opts.Signer,
		Responder:     opts.Responder,
		maxConfigSize: opts.MaxConfigSize,
	}
	svc.api = &api{
		Service:   &svc,
		Responder: opts.Responder,
	}

	// Fetch config version when API requests config version be included in the
	// response
	opts.Responder.Register(tfeapi.IncludeConfig, svc.tfeapi.include)
	// Fetch ingress attributes when API requests ingress attributes be included
	// in the response
	opts.Responder.Register(tfeapi.IncludeIngress, svc.tfeapi.includeIngressAttributes)
	// Resolve authorization requests for config version IDs to a workspace IDs
	opts.Authorizer.RegisterParentResolver(resource.ConfigVersionKind,
		func(ctx context.Context, cvID resource.ID) (resource.ID, error) {
			cvTfeID, ok := cvID.(resource.TfeID)
			if !ok {
				return nil, errors.New("invalid id")
			}
			cv, err := svc.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{
				ID: &cvTfeID,
			})
			if err != nil {
				return nil, err
			}
			return cv.WorkspaceID, nil
		},
	)

	return &svc
}

func (s *Service) AddHandlers(r *mux.Router) {
	s.tfeapi.addHandlers(r)
	s.api.addHandlers(r)
}

func (s *Service) Create(ctx context.Context, workspaceID resource.TfeID, opts CreateOptions) (*ConfigurationVersion, error) {
	subject, err := s.Authorize(ctx, authz.CreateConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		s.Error(err, "constructing configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	if err := s.db.CreateConfigurationVersion(ctx, cv); err != nil {
		s.Error(err, "creating configuration version", "id", cv.ID, "subject", subject)
		return nil, err
	}
	s.V(1).Info("created configuration version", "id", cv.ID, "subject", subject)
	return cv, nil
}

func (s *Service) List(ctx context.Context, workspaceID resource.TfeID, opts ListOptions) (*resource.Page[*ConfigurationVersion], error) {
	subject, err := s.Authorize(ctx, authz.ListConfigurationVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cvl, err := s.db.ListConfigurationVersions(ctx, workspaceID, ListOptions{PageOptions: opts.PageOptions})
	if err != nil {
		s.Error(err, "listing configuration versions")
		return nil, err
	}

	s.V(9).Info("listed configuration versions", "subject", subject)
	return cvl, nil
}

func (s *Service) Get(ctx context.Context, cvID resource.TfeID) (*ConfigurationVersion, error) {
	subject, err := s.Authorize(ctx, authz.GetConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", cvID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved configuration version", "id", cvID, "subject", subject)
	return cv, nil
}

func (s *Service) GetLatest(ctx context.Context, workspaceID resource.TfeID) (*ConfigurationVersion, error) {
	subject, err := s.Authorize(ctx, authz.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		s.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved latest configuration version", "workspace_id", workspaceID, "subject", subject)
	return cv, nil
}

func (s *Service) Delete(ctx context.Context, cvID resource.TfeID) error {
	subject, err := s.Authorize(ctx, authz.DeleteConfigurationVersionAction, cvID)
	if err != nil {
		return err
	}

	err = s.db.DeleteConfigurationVersion(ctx, cvID)
	if err != nil {
		s.Error(err, "deleting configuration version", "id", cvID, "subject", subject)
		return err
	}
	s.V(2).Info("deleted configuration version", "id", cvID, "subject", subject)
	return nil
}
