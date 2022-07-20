package app

import (
	"context"
	"fmt"

	"github.com/leg100/otf"
)

func (a *Application) CreateConfigurationVersion(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionCreateOptions) (*otf.ConfigurationVersion, error) {
	cv, err := otf.NewConfigurationVersion(workspaceID, opts)
	if err != nil {
		a.Error(err, "constructing configuration version", "id", cv.ID())
		return nil, err
	}
	if err := a.db.CreateConfigurationVersion(ctx, cv); err != nil {
		a.Error(err, "creating configuration version", "id", cv.ID())
		return nil, err
	}
	a.V(2).Info("created configuration version", "id", cv.ID())
	return cv, nil
}

func (a *Application) ListConfigurationVersions(ctx context.Context, workspaceID string, opts otf.ConfigurationVersionListOptions) (*otf.ConfigurationVersionList, error) {
	cvl, err := a.db.ListConfigurationVersions(ctx, workspaceID, otf.ConfigurationVersionListOptions{ListOptions: opts.ListOptions})
	if err != nil {
		a.Error(err, "listing configuration versions")
		return nil, err
	}
	a.V(2).Info("listed configuration versions")
	return cvl, nil
}

func (a *Application) GetConfigurationVersion(ctx context.Context, cvID string) (*otf.ConfigurationVersion, error) {
	cv, err := a.db.GetConfigurationVersion(ctx, otf.ConfigurationVersionGetOptions{ID: &cvID})
	if err != nil {
		a.Error(err, "retrieving configuration version", "id", cvID)
		return nil, err
	}
	a.V(2).Info("retrieved configuration version", "id", cvID)
	return cv, nil
}

func (a *Application) GetLatestConfigurationVersion(ctx context.Context, workspaceID string) (*otf.ConfigurationVersion, error) {
	cv, err := a.db.GetConfigurationVersion(ctx, otf.ConfigurationVersionGetOptions{WorkspaceID: &workspaceID})
	if err != nil {
		a.Error(err, "retrieving latest configuration version", "workspace_id", workspaceID)
		return nil, err
	}
	a.V(2).Info("retrieved latest configuration version", "workspace_id", workspaceID)
	return cv, nil
}

// UploadConfig saves a configuration tarball to the db
func (a *Application) UploadConfig(ctx context.Context, cvID string, config []byte) error {
	err := a.db.UploadConfigurationVersion(context.Background(), cvID, func(cv *otf.ConfigurationVersion, uploader otf.ConfigUploader) error {
		return cv.Upload(context.Background(), config, uploader)
	})
	if err != nil {
		a.Error(err, "uploading configuration")
		return err
	}
	if err := a.cache.Set(otf.ConfigVersionCacheKey(cvID), config); err != nil {
		return fmt.Errorf("caching configuration version tarball: %w", err)
	}
	if err != nil {
		a.Error(err, "uploading configuration")
		return err
	}
	a.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

func (a *Application) DownloadConfig(ctx context.Context, cvID string) ([]byte, error) {
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
	a.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return config, nil
}
