// Package releases manages terraform releases.
package releases

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
)

const (
	DefaultTerraformVersion = "1.6.0"
	LatestVersionString     = "latest"
)

type (
	Service struct {
		logr.Logger
		*downloader
		latestChecker

		db     *db
		engine engine.Engine
	}

	Options struct {
		logr.Logger
		*sql.DB

		BinDir string        // destination directory for binaries
		Engine engine.Engine // terraform or tofu
	}
)

func NewService(opts Options) *Service {
	svc := &Service{
		Logger:        opts.Logger,
		db:            &db{opts.DB},
		latestChecker: latestChecker{latestEndpoint},
		downloader:    NewDownloader(opts.Engine, opts.BinDir),
	}
	return svc
}

// StartLatestChecker starts the latest checker go routine, checking the Hashicorp
// API endpoint for a new latest version.
func (s *Service) StartLatestChecker(ctx context.Context) {
	check := func() {
		err := func() error {
			before, checkpoint, err := s.GetLatest(ctx)
			if err != nil {
				return err
			}
			after, err := s.latestChecker.check(checkpoint)
			if err != nil {
				return err
			}
			if after == "" {
				// check was skipped (too early)
				return nil
			}
			// perform sanity check
			if n := semver.Compare(after, before); n < 0 {
				return fmt.Errorf("endpoint returned older version: before: %s; after: %s", before, after)
			}
			// update db (even if version hasn't changed we need to update the
			// checkpoint)
			if err := s.db.updateLatestVersion(ctx, after); err != nil {
				return err
			}
			s.V(1).Info("checked latest terraform version", "before", before, "after", after)
			return nil
		}()
		if err != nil {
			s.Error(err, "checking latest terraform version")
		}
	}
	// check once at startup
	check()
	// ...and check every 5 mins thereafter
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				check()
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

// GetLatest returns the latest terraform version and the time when it was
// fetched; if it has not yet been fetched then the default version is returned
// instead along with zero time.
func (s *Service) GetLatest(ctx context.Context) (string, time.Time, error) {
	latest, checkpoint, err := s.db.getLatest(ctx)
	if errors.Is(err, internal.ErrResourceNotFound) {
		// no latest version has yet been persisted to the database so return
		// the default version instead
		return DefaultTerraformVersion, time.Time{}, nil
	} else if err != nil {
		return "", time.Time{}, err
	}
	return latest, checkpoint, nil
}
