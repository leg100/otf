package configversion

import (
	"context"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/configversion/source"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/resource"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/tfeapi"
	"github.com/leg100/surl/v2"
)

type (
	Service struct {
		logr.Logger
		*authz.Authorizer
		*source.IconDB

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
		IconDB:     source.NewIconDB(),
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

	// Provide a means of looking up a config version's parent workspace.
	opts.Authorizer.RegisterParentResolver(resource.ConfigVersionKind,
		func(ctx context.Context, cvID resource.ID) (resource.ID, error) {
			// NOTE: we look up directly in the database rather than via
			// service call to avoid a recursion loop.
			cv, err := svc.db.get(ctx, cvID)
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

	cv := NewConfigurationVersion(workspaceID, opts)
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

// ListOlderThan lists configs created before t. Implements resource.deleterClient.
func (s *Service) ListOlderThan(ctx context.Context, t time.Time) ([]*ConfigurationVersion, error) {
	return s.db.listOlderThan(ctx, t)
}

func (s *Service) Get(ctx context.Context, id resource.TfeID) (*ConfigurationVersion, error) {
	subject, err := s.Authorize(ctx, authz.GetConfigurationVersionAction, id)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.get(ctx, id)
	if err != nil {
		s.Error(err, "retrieving configuration version", "id", id, "subject", subject)
		return nil, err
	}
	s.V(9).Info("retrieved configuration version", "id", id, "subject", subject)
	return cv, nil
}

func (s *Service) GetLatest(ctx context.Context, workspaceID resource.TfeID) (*ConfigurationVersion, error) {
	subject, err := s.Authorize(ctx, authz.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := s.db.getLatest(ctx, workspaceID)
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
