package engine

import (
	"context"
	"errors"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
)

type (
	// Alias service to permit embedding it with other services in a struct
	// without a name clash.
	EngineService = Service

	Service struct {
		logger logr.Logger
		db     DB
	}

	Options struct {
		Logger logr.Logger
		DB     *sql.DB
		BinDir string // destination directory for binaries
	}

	DB interface {
		getLatest(ctx context.Context, engine string) (string, time.Time, error)
		updateLatestVersion(ctx context.Context, engine, v string) error
	}
)

func NewService(opts Options) *Service {
	return &Service{
		logger: opts.Logger,
		db:     &db{opts.DB},
	}
}

// GetLatest returns the latest engine version and the time when it was
// fetched; if it has not yet been fetched then the default version is returned
// instead along with zero time.
func (s *Service) GetLatest(ctx context.Context, engine Kind) (string, time.Time, error) {
	latest, checkpoint, err := s.db.getLatest(ctx, engine.String())
	if errors.Is(err, internal.ErrResourceNotFound) {
		// no latest version has yet been persisted to the database so return
		// the default version instead
		return engine.DefaultVersion(), time.Time{}, nil
	} else if err != nil {
		return "", time.Time{}, err
	}
	return latest, checkpoint, nil
}

func (s *Service) UpdateLatestVersion(ctx context.Context, engine, version string) error {
	return s.db.updateLatestVersion(ctx, engine, version)
}
