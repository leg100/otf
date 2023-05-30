package run

import (
	"context"
	"fmt"

	"github.com/leg100/otf/internal/rbac"
)

func lockFileCacheKey(runID string) string {
	return fmt.Sprintf("%s.terraform.lock.hcl", runID)
}

type lockFileService interface {
	// GetLockFile returns the lock file for the run.
	GetLockFile(ctx context.Context, runID string) ([]byte, error)
	// UploadLockFile persists the lock file for a run.
	UploadLockFile(ctx context.Context, runID string, plan []byte) error
}

// getLockFile returns the lock file for the run.
func (s *service) GetLockFile(ctx context.Context, runID string) ([]byte, error) {
	subject, err := s.CanAccess(ctx, rbac.GetLockFileAction, runID)
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
func (s *service) UploadLockFile(ctx context.Context, runID string, file []byte) error {
	subject, err := s.CanAccess(ctx, rbac.UploadLockFileAction, runID)
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
