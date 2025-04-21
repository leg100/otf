package releases

import (
	"context"
	"net/url"
	"time"

	"github.com/leg100/otf/internal/engine"
)

type testEngine struct {
	engine.Engine
	u             *url.URL
	latestVersion string
}

func (e *testEngine) SourceURL(version string) *url.URL { return e.u }
func (e *testEngine) String() string                    { return "terraform" }
func (e *testEngine) GetLatestVersion(context.Context) (string, error) {
	return e.latestVersion, nil
}

type testDB struct {
	DB
	lastCheck     time.Time
	currentLatest string
}

func (db *testDB) updateLatestVersion(ctx context.Context, engine, v string) error {
	return nil
}

func (db *testDB) getLatest(ctx context.Context, engine string) (string, time.Time, error) {
	return db.currentLatest, db.lastCheck, nil
}
