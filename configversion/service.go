package configversion

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/surl"
)

type Service struct {
	app app
	*api
}

func NewService(opts Options) *Service {
	app := &Application{
		Authorizer: opts.Authorizer,
		cache:      opts.Cache,
		db:         newPGDB(opts.Database),
		Logger:     opts.Logger,
	}
	return &Service{
		app: app,
		api: newAPI(apiOptions{app, opts.MaxUploadSize, opts.Signer}),
	}
}

type Options struct {
	otf.Authorizer
	otf.Cache
	otf.Database
	*surl.Signer
	logr.Logger
	MaxUploadSize int64
}

func (s *Service) CreateConfigurationVersion(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (otf.ConfigurationVersion, error) {
	return s.app.create(ctx, workspaceID, opts)
}

func (s *Service) CloneConfigurationVersion(ctx context.Context, cvID string, opts otf.ConfigurationVersionCreateOptions) (otf.ConfigurationVersion, error) {
	return s.app.clone(ctx, cvID, opts)
}

func (s *Service) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	return s.app.list(ctx, workspaceID, opts)
}

func (s *Service) GetConfigurationVersion(ctx context.Context, cvID string) (otf.ConfigurationVersion, error) {
	return s.app.get(ctx, cvID)
}

func (s *Service) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (otf.ConfigurationVersion, error) {
	return s.app.getLatest(ctx, workspaceID)
}
