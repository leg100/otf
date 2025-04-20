// Package releases manages engine releases.
package releases

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
)

const LatestVersionString = "latest"

type (
	Service struct {
		logr.Logger
		*downloader

		db     DB
		engine Engine
	}

	Options struct {
		logr.Logger
		*sql.DB

		BinDir string // destination directory for binaries
		Engine Engine // terraform or tofu
	}

	DB interface {
		getLatest(ctx context.Context) (string, time.Time, error)
		updateLatestVersion(ctx context.Context, v string) error
	}

	Engine interface {
		String() string
		DefaultVersion() string
		GetLatestVersion(context.Context) (string, error)
		SourceURL(version string) *url.URL
	}
)

func NewService(opts Options) *Service {
	svc := &Service{
		Logger:     opts.Logger,
		db:         &db{opts.DB, opts.Engine},
		downloader: NewDownloader(opts.Engine, opts.BinDir),
		engine:     opts.Engine,
	}
	return svc
}

// StartLatestChecker starts the latest checker go routine, checking the Hashicorp
// API endpoint for a new latest version.
func (s *Service) StartLatestChecker(ctx context.Context) {
	// check once at startup
	s.checkAndLog(ctx)
	// ...and check every 5 mins thereafter
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				s.checkAndLog(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) checkAndLog(ctx context.Context) {
	before, after, err := s.check(ctx)
	if err != nil {
		s.Error(err, "checking latest engine version", "engine", s.engine)
		return
	}
	// update db (even if version hasn't changed we need to update the
	// checkpoint)
	if err := s.db.updateLatestVersion(ctx, after); err != nil {
		s.Error(err, "checking latest engine version", "engine", s.engine)
		return
	}
	s.V(1).Info("checked latest engine version", "engine", s.engine, "before", before, "after", after)
}

func (s *Service) check(ctx context.Context) (before string, after string, err error) {
	// get current latest version stored in db
	before, checkpoint, err := s.GetLatest(ctx)
	if err != nil {
		return "", "", err
	}
	// skip check if already checked within last 24 hours
	if checkpoint.After(time.Now().Add(-24 * time.Hour)) {
		return "", "", nil
	}
	// get latest version from engine's internet endpoint
	after, err = s.engine.GetLatestVersion(ctx)
	if err != nil {
		return "", "", err
	}
	// perform sanity check
	if n := semver.Compare(after, before); n < 0 {
		return "", "", fmt.Errorf("endpoint returned older version: before: %s; after: %s", before, after)
	}
	return before, after, nil
}

// GetLatest returns the latest engine version and the time when it was
// fetched; if it has not yet been fetched then the default version is returned
// instead along with zero time.
func (s *Service) GetLatest(ctx context.Context) (string, time.Time, error) {
	latest, checkpoint, err := s.db.getLatest(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// no latest version has yet been persisted to the database so return
		// the default version instead
		return s.engine.DefaultVersion(), time.Time{}, nil
	} else if err != nil {
		return "", time.Time{}, err
	}
	return latest, checkpoint, nil
}
