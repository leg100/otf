package run

import (
	"context"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

// GetLockFile returns the lock file for the run.
func (s *Service) GetLockFile(ctx context.Context, runID resource.ID) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.GetLockFileAction, runID)
	if err != nil {
		return nil, err
	}

	file, err := s.db.GetLockFile(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving lock file", "id", runID, "subject", subject)
		return nil, err
	}

	return file, nil
}

// UploadLockFile persists the lock file for a run.
func (s *Service) UploadLockFile(ctx context.Context, runID resource.ID, file []byte) error {
	subject, err := s.Authorize(ctx, authz.UploadLockFileAction, runID)
	if err != nil {
		return err
	}

	if err := s.db.SetLockFile(ctx, runID, file); err != nil {
		s.Error(err, "uploading lock file", "id", runID, "subject", subject)
		return err
	}
	s.V(1).Info("uploaded lock file", "id", runID)

	return nil
}
