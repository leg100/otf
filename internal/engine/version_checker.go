package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/leg100/otf/internal/semver"
)

type VersionChecker struct {
	Logger logr.Logger
	Client VersionCheckerClient
}

type VersionCheckerClient interface {
	GetLatest(ctx context.Context, engine *Engine) (string, time.Time, error)
	UpdateLatestVersion(ctx context.Context, engine *Engine, v string) error
}

// Start the latest checker go routine, checking the latest version of each
// engine on a regular interval.
func (s *VersionChecker) Start(ctx context.Context) error {
	// check once at startup
	s.checkAndUpdateAll(ctx)
	// ...and check every 5 mins thereafter
	ticker := time.NewTicker(5 * time.Minute)
	for {
		select {
		case <-ticker.C:
			s.checkAndUpdateAll(ctx)
		case <-ctx.Done():
			return nil
		}
	}
}

func (s *VersionChecker) checkAndUpdateAll(ctx context.Context) {
	for _, engine := range Engines() {
		result, err := s.check(ctx, engine, time.Now())
		if err != nil {
			s.Logger.Error(err, "checking latest engine version", "engine", engine)
			return
		}
		if !result.skipped {
			// update db (even if version hasn't changed we need to update the
			// checkpoint)
			if err := s.Client.UpdateLatestVersion(ctx, engine, result.after); err != nil {
				s.Logger.Error(err, "persisting db with latest engine version", "engine", engine)
				return
			}
		}
		s.Logger.V(3).Info(result.message, "engine", engine, "check", result)
	}
}

func (s *VersionChecker) check(ctx context.Context, engine *Engine, now time.Time) (result checkResult, err error) {
	// get current latest version stored in db
	before, checkpoint, err := s.Client.GetLatest(ctx, engine)
	if err != nil {
		return checkResult{}, err
	}
	// skip check if already checked within last 24 hours
	if checkpoint.After(now.Add(-24 * time.Hour)) {
		return checkResult{
			before:         before,
			skipped:        true,
			nextCheckpoint: checkpoint,
			message:        "skipped latest engine version check",
		}, nil
	}
	// get latest version from engine's internet endpoint
	after, err := engine.LatestVersionGetter.Get(ctx)
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
