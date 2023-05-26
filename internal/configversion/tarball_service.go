package configversion

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
)

// cacheKey generates a key for caching config tarballs
func cacheKey(cvID string) string {
	return fmt.Sprintf("%s.tar.gz", cvID)
}

// upload saves a configuration tarball to the db
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *service) UploadConfig(ctx context.Context, cvID string, config []byte) error {
	err := s.db.UploadConfigurationVersion(ctx, cvID, func(cv *ConfigurationVersion, uploader ConfigUploader) error {
		return cv.Upload(ctx, config, uploader)
	})
	if err != nil {
		s.Error(err, "uploading configuration")
		return err
	}
	if err := s.cache.Set(cacheKey(cvID), config); err != nil {
		s.Error(err, "caching configuration version tarball")
	}
	s.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

// download retrieves a tarball from the db
func (s *service) DownloadConfig(ctx context.Context, cvID string) ([]byte, error) {
	subject, err := s.canAccess(ctx, rbac.DownloadConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	if config, err := s.cache.Get(cacheKey(cvID)); err == nil {
		return config, nil
	}
	config, err := s.db.GetConfig(ctx, cvID)
	if err != nil {
		return nil, err
	}
	if err := s.cache.Set(cacheKey(cvID), config); err != nil {
		s.Error(err, "caching configuration version tarball")
	}
	s.V(9).Info("downloaded configuration", "id", cvID, "bytes", len(config), "subject", subject)
	return config, nil
}
