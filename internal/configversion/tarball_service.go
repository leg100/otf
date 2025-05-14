package configversion

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

// cacheKey generates a key for caching config tarballs
func cacheKey(cvID resource.TfeID) string {
	return fmt.Sprintf("%s.tar.gz", cvID)
}

// UploadConfig saves a configuration tarball to the db
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) UploadConfig(ctx context.Context, cvID resource.TfeID, config []byte) error {
	err := s.db.UploadConfigurationVersion(ctx, cvID, config)
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

// DownloadConfig retrieves a tarball from the db
func (s *Service) DownloadConfig(ctx context.Context, cvID resource.TfeID) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.DownloadConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	if config, err := s.cache.Get(cacheKey(cvID)); err == nil {
		return config, nil
	}
	config, err := s.db.GetConfig(ctx, cvID)
	if err != nil {
		s.Error(err, "downloading configuration", "id", cvID, "subject", subject)
		return nil, err
	}
	if err := s.cache.Set(cacheKey(cvID), config); err != nil {
		s.Error(err, "caching configuration version tarball")
	}
	s.V(9).Info("downloaded configuration", "id", cvID, "bytes", len(config), "subject", subject)
	return config, nil
}
