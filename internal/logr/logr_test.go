package logr

import (
	"bytes"
	"errors"
	"testing"

	"log/slog"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name string
		min  slog.Leveler
		log  func(logger logr.Logger)
		want string
	}{
		{
			"info",
			slog.LevelInfo,
			func(logger logr.Logger) {
				logger.Info("something", "foo", "bar")
			},
			"level=INFO msg=something foo=bar\n",
		},
		{
			"error",
			slog.LevelInfo,
			func(logger logr.Logger) {
				logger.Error(errors.New("woops"), "spilt me beer", "foo", "bar")
			},
			"level=ERROR msg=\"spilt me beer\" error=woops foo=bar\n",
		},
		{
			"debug",
			slog.LevelDebug,
			func(logger logr.Logger) {
				logger.V(1).Info("something", "foo", "bar")
			},
			"level=DEBUG msg=something foo=bar\n",
		},
		{
			"debug",
			slog.Level(-5),
			func(logger logr.Logger) {
				logger.V(1).Info("something", "foo", "bar")
			},
			"level=DEBUG msg=something foo=bar\n",
		},
		{
			"hide debug",
			slog.LevelInfo,
			func(logger logr.Logger) {
				logger.V(1).Info("should not see this", "foo", "bar")
			},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got bytes.Buffer
			logger := logr.New(newLogSink(slog.NewTextHandler(&got, newTestOptions(tt.min))))
			tt.log(logger)
			assert.Equal(t, tt.want, got.String())
		})
	}
}

func newTestOptions(min slog.Leveler) *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: min,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Remove time.
			if a.Key == slog.TimeKey && len(groups) == 0 {
				return slog.Attr{}
			}
			return a
		},
	}
}
