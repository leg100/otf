// Package releases manages terraform releases.
package releases

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/gorilla/mux"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/sql"
	"github.com/leg100/otf/internal/sql/pggen"
)

const DefaultTerraformVersion = "1.5.2"

type (
	ReleasesService = Service

	Service interface {
		// Download a terraform release with the given version and log progress
		// updates to logger. Once complete, the path to the release executable
		// is returned.
		Download(ctx context.Context, version string, logger io.Writer) (string, error)
		// getLatest returns the latest version of terraform along with the
		// time when the latest version was last determined.
		getLatest(ctx context.Context) (string, time.Time, error)
	}
	service struct {
		logr.Logger
		*sql.DB
		*downloader
		*api

		latestChecker
	}
	Options struct {
		logr.Logger
		*sql.DB

		TerraformBinDir string // destination directory for terraform binaries
	}
)

func NewService(opts Options) *service {
	svc := &service{
		Logger:        opts.Logger,
		DB:            opts.DB,
		downloader:    newDownloader(opts.TerraformBinDir),
		latestChecker: latestChecker{latestEndpoint},
	}
	svc.api = &api{svc}

	return svc
}

func (s *service) AddHandlers(r *mux.Router) {
	s.api.addHandlers(r)
}

func (s *service) Download(ctx context.Context, version string, logger io.Writer) (string, error) {
	if version == "latest" {
		var err error
		version, _, err = s.getLatest(ctx)
		if err != nil {
			return "", err
		}
	}
	return s.downloader.Download(ctx, version, logger)
}

func (s *service) StartLatestChecker(ctx context.Context) {
	check := func() {
		err := func() error {
			current, checkpoint, err := s.getLatest(ctx)
			if err != nil {
				return err
			}
			newer, checked, err := s.latestChecker.check(checkpoint, current)
			if err != nil {
				return err
			}
			return s.Tx(ctx, func(ctx context.Context, q pggen.Querier) error {
				if checked {
					// endpoint was checked; update checkpoint in database
					_, err := q.UpdateLatestTerraformVersionCheckpoint(ctx)
					if err != nil {
						return err
					}
					s.Info("checked latest terraform version", "before", current, "after", newer)
				}
				if newer != "" {
					// newer version found; update version in database
					_, err := q.UpdateLatestTerraformVersion(ctx, sql.String(newer))
					if err != nil {
						return err
					}
					s.Info("updated latest terraform version", "before", current, "after", newer)
				}
				return nil
			})
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

// getLatest returns the latest terraform version and the time when it was
// fetched; if it has not yet been fetched then the default version is returned
// instead along with zero time.
func (s *service) getLatest(ctx context.Context) (string, time.Time, error) {
	row, err := s.Conn(ctx).FindLatestTerraformVersion(ctx)
	if errors.Is(sql.Error(err), internal.ErrResourceNotFound) {
		// no latest version has yet been persisted to the database so return
		// the default version instead
		return DefaultTerraformVersion, time.Time{}, nil
	} else if err != nil {
		return "", time.Time{}, err
	}
	return row.Version.String, row.Checkpoint.Time, nil
}
