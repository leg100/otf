package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/authz"
	"github.com/leg100/otf/internal/resource"
)

func lockFileCacheKey(runID resource.ID) string {
	return fmt.Sprintf("%s.terraform.lock.hcl", runID)
}

// GetLockFile returns the lock file for the run.
func (s *Service) GetLockFile(ctx context.Context, runID resource.ID) ([]byte, error) {
	subject, err := s.Authorize(ctx, authz.GetLockFileAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return nil, err
	}

	if plan, err := s.cache.Get(lockFileCacheKey(runID)); err == nil {
		return plan, nil
	}
	// cache miss; retrieve from db
	file, err := s.db.GetLockFile(ctx, runID)
	if err != nil {
		s.Error(err, "retrieving lock file", "id", runID, "subject", subject)
		return nil, err
	}

	// cache lock file before returning
	if err := s.cache.Set(lockFileCacheKey(runID), file); err != nil {
		s.Error(err, "caching lock file")
	}
	return file, nil
}

// UploadLockFile persists the lock file for a run.
func (s *Service) UploadLockFile(ctx context.Context, runID resource.ID, file []byte) error {
	subject, err := s.Authorize(ctx, authz.UploadLockFileAction, &authz.AccessRequest{ID: &runID})
	if err != nil {
		return err
	}

	if err := s.db.SetLockFile(ctx, runID, file); err != nil {
		s.Error(err, "uploading lock file", "id", runID, "subject", subject)
		return err
	}
	s.V(1).Info("uploaded lock file", "id", runID)

	// cache lock file before returning
	if err := s.cache.Set(lockFileCacheKey(runID), file); err != nil {
		s.Error(err, "caching lock file")
	}
	return nil
}
