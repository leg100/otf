package logr

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
)

type (
	Config struct {
		Verbosity int    `name:"v" help:"Logging level"`
		Format    string `name:"log-format" help:"Logging format: text or json"`
	}

	Format string
)

const (
	DefaultFormat Format = "default"
	TextFormat    Format = "text"
	JSONFormat    Format = "json"
)

// NewConfigFromFlags adds flags to the given flagset, and, after the
// flagset is parsed by the caller, the flags populate the returned logger
// config.
func NewConfigFromFlags(flags *pflag.FlagSet) *Config {
	cfg := Config{}
	flags.IntVarP(&cfg.Verbosity, "v", "v", 0, "Logging level")
	flags.StringVar(&cfg.Format, "log-format", string(DefaultFormat), "Logging format: text or json")
	return &cfg
}

// New constructs a new logger that satisfies the logr interface
func New(cfg *Config) (logr.Logger, error) {
	var h slog.Handler
	level := toSlogLevel(cfg.Verbosity)

	switch Format(cfg.Format) {
	case DefaultFormat:
		h = NewLevelHandler(level, slog.Default().Handler())
	case TextFormat:
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case JSONFormat:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	default:
		return logr.Logger{}, fmt.Errorf("unrecognised logging format: %s", cfg.Format)
	}
	return logr.New(newLogSink(h)), nil
}

// toSlogLevel converts a logr v-level to a slog level.
func toSlogLevel(verbosity int) slog.Level {
	if verbosity <= 0 {
		return slog.LevelInfo
	}
	return slog.Level(-4 - (verbosity - 1))
}
