package configversion

import (
	"context"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

// UploadConfig saves a configuration tarball to the db
//
// NOTE: unauthenticated - access granted only via signed URL
func (s *Service) UploadConfig(ctx context.Context, cvID resource.ID, config []byte) error {
	err := s.db.UploadConfigurationVersion(ctx, cvID, config)
	if err != nil {
		s.Error(err, "uploading configuration")
		return err
	}
	s.V(2).Info("uploaded configuration", "id", cvID, "bytes", len(config))
	return nil
}

// DownloadConfig retrieves a tarball from the db
func (s *Service) DownloadConfig(ctx context.Context, cvID resource.ID) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.DownloadConfigurationVersionAction, cvID)
	if err != nil {
		return nil, err
	}

	config, err := s.db.GetConfig(ctx, cvID)
	if err != nil {
		s.Error(err, "downloading configuration", "id", cvID, "subject", subject)
		return nil, err
	}
	s.V(9).Info("downloaded configuration", "id", cvID, "bytes", len(config), "subject", subject)
	return config, nil
}
