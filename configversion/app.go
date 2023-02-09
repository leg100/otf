package configversion

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/leg100/otf"
	"github.com/leg100/otf/rbac"
	"github.com/leg100/surl"
)

type app interface {
	CreateConfigurationVersion(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
	// CloneConfigurationVersion creates a new configuration version using the
	// config tarball of an existing configuration version.
	CloneConfigurationVersion(ctx context.Context, cvID string, opts otf.ConfigurationVersionCreateOptions) (*ConfigurationVersion, error)
	GetConfigurationVersion(ctx context.Context, id string) (*ConfigurationVersion, error)
	GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*ConfigurationVersion, error)
	ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error)

	// Upload handles verification and upload of the config tarball, updating
	// the config version upon success or failure.
	UploadConfig(ctx context.Context, id string, config []byte) error

	// Download retrieves the config tarball for the given config version ID.
	DownloadConfig(ctx context.Context, id string) ([]byte, error)
}

type Application struct {
	otf.Authorizer
	logr.Logger

	db    db
	cache otf.Cache
	*handlers
}

func NewApplication(opts ApplicationOptions) *Application {
	app := &Application{
		Authorizer: opts.Authorizer,
		db:         newPGDB(opts.Database),
		Logger:     opts.Logger,
	}
	app.handlers = newHandlers(handlersOptions{app, opts.MaxUploadSize, opts.Signer})
	return app
}

type ApplicationOptions struct {
	otf.Authorizer
	otf.Database
	*surl.Signer
	logr.Logger
	MaxUploadSize int64
}

func (a *Application) CreateConfigurationVersion(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (otf.ConfigurationVersion, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.CreateConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing configuration version", "id", cv.ID(), "subject", subject)
		return nil, err
	}
	if err := a.db.CreateConfigurationVersion(ctx, cv); err != nil {
		a.Error(err, "creating configuration version", "id", cv.ID(), "subject", subject)
		return nil, err
	}
	a.V(2).Info("created configuration version", "id", cv.ID(), "subject", subject)
	return cv, nil
}

func (a *Application) CloneConfigurationVersion(ctx context.Context, cvID string, opts otf.ConfigurationVersionCreateOptions) (otf.ConfigurationVersion, error) {
	cv, err := a.GetConfigurationVersion(ctx, cvID)
	if err != nil {
		return nil, err
	}

	cv, err = a.CreateConfigurationVersion(ctx, cv.WorkspaceID(), opts)
	if err != nil {
		return nil, err
	}

	config, err := a.DownloadConfig(ctx, cvID)
	if err != nil {
		return nil, err
	}

	if err := a.UploadConfig(ctx, cv.ID(), config); err != nil {
		return nil, err
	}

	return cv, nil
}

func (a *Application) ListConfigurationVersions(ctx context.Context, workspaceID string, opts ConfigurationVersionListOptions) (*ConfigurationVersionList, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.ListConfigurationVersionsAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cvl, err := a.db.ListConfigurationVersions(ctx, workspaceID, ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		a.Error(err, "listing configuration versions", "subject", subject)
		return nil, err
	}
	a.V(2).Info("listed configuration versions", "subject", subject)
	return cvl, nil
}

func (a *Application) GetConfigurationVersion(ctx context.Context, cvID string) (otf.ConfigurationVersion, error) {
	subject, err := a.CanAccessConfigurationVersion(ctx, rbac.GetConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	cv, err := a.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		a.Error(err, "retrieving configuration version", "id", cvID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved configuration version", "id", cvID, "subject", subject)
	return cv, nil
}

func (a *Application) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (otf.ConfigurationVersion, error) {
	subject, err := a.CanAccessWorkspaceByID(ctx, rbac.GetConfigurationVersionAction, workspaceID)
	if err != nil {
		return nil, err
	}

	cv, err := a.db.GetConfigurationVersion(ctx, ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID, "subject", subject)
		return nil, err
	}
	a.V(2).Info("retrieved latest configuration version", "workspace_id", workspaceID, "subject", subject)
	return cv, nil
}

// UploadConfig saves a configuration tarball to the db
//
// NOTE: unauthenticated - access granted only via signed URL
func (a *Application) UploadConfig(ctx context.Context, cvID string, config []byte) error {
	err := a.db.UploadConfigurationVersion(ctx, cvID, func(cv *ConfigurationVersion, uploader ConfigUploader) error {
		return cv.Upload(ctx, config, uploader)
	})
	if err != nil {
		a.Error(err, "uploading configuration")
		return err
	}
	if err := a.cache.Set(otf.ConfigVersionCacheKey(cvID), config); err != nil {
		return fmt.Errorf("caching configuration version tarball: %w", err)
	}
	a.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

func (a *Application) DownloadConfig(ctx context.Context, cvID string) ([]byte, error) {
	subject, err := a.CanAccessConfigurationVersion(ctx, rbac.DownloadConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	if config, err := a.cache.Get(otf.ConfigVersionCacheKey(cvID)); err == nil {
		return config, nil
	}
	config, err := a.db.GetConfig(context.Background(), cvID)
	if err != nil {
		return nil, err
	}
	if err := a.cache.Set(otf.ConfigVersionCacheKey(cvID), config); err != nil {
		return nil, fmt.Errorf("caching configuration version tarball: %w", err)
	}
	a.V(2).Info("downloaded configuration", "id", cvID, "bytes", len(config), "subject", subject)
	return config, nil
}
