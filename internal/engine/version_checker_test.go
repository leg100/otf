package engine

import (
	"context"
	"testing"
	"time"

	"github.com/leg100/otf/internal/logr"
	"github.com/stretchr/testify/assert"
)

type fakeVersionCheckerClient struct {
	currentLatest string
	lastCheck     time.Time
}

func (c *fakeVersionCheckerClient) GetLatest(ctx context.Context, engine *Engine) (string, time.Time, error) {
	return c.currentLatest, c.lastCheck, nil
}

func (c *fakeVersionCheckerClient) UpdateLatestVersion(ctx context.Context, engine *Engine, v string) error {
	return nil
}

type fakeLatestVersionGetter struct {
	v string
}

func (f *fakeLatestVersionGetter) Get(context.Context) (string, error) {
	return f.v, nil
}

func Test_check(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name          string
		currentLatest string
		newLatest     string
		lastCheck     time.Time
		want          checkResult
	}{
		{
			"perform check and expect newer version",
			"1.2.3",
			"1.2.4",
			now.Add(-time.Hour * 48),
			checkResult{before: "1.2.3", after: "1.2.4", nextCheckpoint: now.Add(time.Hour * 24), message: "updated latest engine version"},
		},
		{
			"perform check and expect same version",
			"1.2.3",
			"1.2.3",
			now.Add(-time.Hour * 48),
			checkResult{before: "1.2.3", after: "1.2.3", nextCheckpoint: now.Add(time.Hour * 24), message: "updated latest engine version"},
		},
		{
			"skip check because previous check performed an hour ago",
			"",
			"",
			now.Add(-time.Hour),
			checkResult{skipped: true, nextCheckpoint: now.Add(time.Hour * 23), message: "skipped latest engine version check"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &VersionChecker{
				Logger: logr.Discard(),
				Client: &fakeVersionCheckerClient{
					currentLatest: tt.currentLatest,
					lastCheck:     tt.lastCheck,
				},
			}
			engine := &Engine{
				Name: "test",
				LatestVersionGetter: &fakeLatestVersionGetter{
					v: tt.newLatest,
				},
			}
			got, err := c.check(t.Context(), engine, now)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
