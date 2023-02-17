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

func (s *Service) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	// convert from list of concrete CVs to list of interface CVs
	from, err := s.app.list(ctx, workspaceID, opts)
	if err != nil {
		return nil, err
	}
	to := otf.ConfigurationVersionList{
		Pagination: from.Pagination,
	}
	for _, i := range from.Items {
		to.Items = append(to.Items, i)
	}
	return &to, nil
}

func (s *Service) GetConfigurationVersion(ctx context.Context, cvID string) (otf.ConfigurationVersion, error) {
	return s.app.get(ctx, cvID)
}

func (s *Service) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (otf.ConfigurationVersion, error) {
	return s.app.getLatest(ctx, workspaceID)
}

func (s *Service) DeleteConfigurationVersion(ctx context.Context, cvID string) error {
	return s.app.delete(ctx, cvID)
}

func (s *Service) UploadConfig(ctx context.Context, workspaceID string, config []byte) error {
	return s.app.upload(ctx, workspaceID, config)
}

func (s *Service) DownloadConfig(ctx context.Context, workspaceID string) ([]byte, error) {
	return s.app.download(ctx, workspaceID)
}
