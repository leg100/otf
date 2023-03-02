package configversion

import (
	"context"
	"fmt"

	"github.com/leg100/otf/rbac"
)

// cacheKey generates a key for caching config tarballs
func cacheKey(cvID string) string {
	return fmt.Sprintf("%s.tar.gz", cvID)
}

// upload saves a configuration tarball to the db
//
// NOTE: unauthenticated - access granted only via signed URL
func (a *Service) upload(ctx context.Context, cvID string, config []byte) error {
	err := a.db.UploadConfigurationVersion(ctx, cvID, func(cv *ConfigurationVersion, uploader ConfigUploader) error {
		return cv.Upload(ctx, config, uploader)
	})
	if err != nil {
		a.Error(err, "uploading configuration")
		return err
	}
	if err := a.cache.Set(cacheKey(cvID), config); err != nil {
		return fmt.Errorf("caching configuration version tarball: %w", err)
	}
	a.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

// download retrieves a tarball from the db
func (a *Service) download(ctx context.Context, cvID string) ([]byte, error) {
	subject, err := a.canAccess(ctx, rbac.DownloadConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	if config, err := a.cache.Get(cacheKey(cvID)); err == nil {
		return config, nil
	}
	config, err := a.db.GetConfig(ctx, cvID)
	if err != nil {
		return nil, err
	}
	if err := a.cache.Set(cacheKey(cvID), config); err != nil {
		return nil, fmt.Errorf("caching configuration version tarball: %w", err)
	}
	a.V(2).Info("downloaded configuration", "id", cvID, "bytes", len(config), "subject", subject)
	return config, nil
}
