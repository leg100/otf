// Package releases manages engine releases.
package releases

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/engine"
	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/semver"
	"github.com/leg100/otf/internal/sql"
)

const LatestVersionString = "latest"

type (
	Service struct {
		logr.Logger

		db DB
	}

	Options struct {
		logr.Logger
		*sql.DB

		BinDir string // destination directory for binaries
	}

	DB interface {
		getLatest(ctx context.Context, engine string) (string, time.Time, error)
		updateLatestVersion(ctx context.Context, engine, v string) error
	}

	Engine interface {
		String() string
		DefaultVersion() string
		GetLatestVersion(context.Context) (string, error)
		SourceURL(version string) *url.URL
	}
)

func NewService(opts Options) *Service {
	return &Service{
		Logger: opts.Logger,
		db:     &db{opts.DB},
	}
}

// StartLatestChecker starts the latest checker go routine, checking the latest
// version of each engine on a regular interval.
func (s *Service) StartLatestChecker(ctx context.Context) {
	// check once at startup
	s.checkAndUpdate(ctx)
	// ...and check every 5 mins thereafter
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ticker.C:
				s.checkAndUpdate(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *Service) checkAndUpdate(ctx context.Context) {
	for _, engine := range engine.Engines() {
		result, err := s.check(ctx, engine, time.Now())
		if err != nil {
			s.Error(err, "checking latest engine version", "engine", engine)
			return
		}
		if !result.skipped {
			// update db (even if version hasn't changed we need to update the
			// checkpoint)
			if err := s.db.updateLatestVersion(ctx, engine.String(), result.after); err != nil {
				s.Error(err, "persisting db with latest engine version", "engine", engine)
				return
			}
		}
		s.V(3).Info(result.message, "engine", engine, "check", result)
	}
}

// checkResult is the result of the latest version check for an engine.
type checkResult struct {
	before, after  string
	skipped        bool
	nextCheckpoint time.Time
	message        string
}

func (r checkResult) LogValue() slog.Value {
	attrs := []slog.Attr{
		slog.Time("next_check_at", r.nextCheckpoint),
	}
	if r.skipped {
		attrs = append(attrs, slog.String("skip_reason", "check not due yet"))
		attrs = append(attrs, slog.String("current", r.before))
	} else {
		attrs = append(attrs, slog.String("before", r.before))
		attrs = append(attrs, slog.String("after", r.after))
	}
	return slog.GroupValue(attrs...)
}

func (s *Service) check(ctx context.Context, engine Engine, now time.Time) (result checkResult, err error) {
	// get current latest version stored in db
	before, checkpoint, err := s.GetLatest(ctx, engine)
	if err != nil {
		return checkResult{}, err
	}
	// skip check if already checked within last 24 hours
	if checkpoint.After(now.Add(-24 * time.Hour)) {
		return checkResult{
			before:         before,
			skipped:        true,
			nextCheckpoint: checkpoint.Add(24 * time.Hour),
			message:        "skipped latest engine version check",
		}, nil
	}
	// get latest version from engine's internet endpoint
	after, err := engine.GetLatestVersion(ctx)
	if err != nil {
		return checkResult{}, err
	}
	// perform sanity check
	if n := semver.Compare(after, before); n < 0 {
		return checkResult{}, fmt.Errorf("endpoint returned older version: before: %s; after: %s", before, after)
	}
	return checkResult{
		before:         before,
		after:          after,
		nextCheckpoint: now.Add(24 * time.Hour),
		message:        "updated latest engine version",
	}, nil
}

// GetLatest returns the latest engine version and the time when it was
// fetched; if it has not yet been fetched then the default version is returned
// instead along with zero time.
func (s *Service) GetLatest(ctx context.Context, engine Engine) (string, time.Time, error) {
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
