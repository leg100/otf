package logr

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/go-logr/logr"
	"github.com/spf13/pflag"
)

const (
	DefaultFormat Format = "default"
	TextFormat    Format = "text"
	JSONFormat    Format = "json"
)

type (
	// Logger wraps the upstream logr logger, adding further functionality.
	Logger struct {
		logr.Logger
		Config
	}

	Config struct {
		Verbosity int
		Format    string
	}

	Format string
)

// RegisterFlags adds flags to the given flagset, and, after the
// flagset is parsed by the caller, the flags populate the returned logger
// config.
func RegisterFlags(flags *pflag.FlagSet, cfg *Config) {
	flags.IntVarP(&cfg.Verbosity, "v", "v", 0, "Logging level")
	flags.StringVar(&cfg.Format, "log-format", string(DefaultFormat), "Logging format: text or json")
}

// New constructs a new logger that satisfies the logr interface
func New(cfg Config) (Logger, error) {
	var h slog.Handler
	level := toSlogLevel(cfg.Verbosity)

	switch Format(cfg.Format) {
	case DefaultFormat:
		h = NewLevelHandler(level, slog.Default().Handler())
	case TextFormat:
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case JSONFormat:
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	case "":
		// Empty string for format means discard all logs. Undocumented.
		return Discard(), nil
	default:
		return Logger{}, fmt.Errorf("unrecognised logging format: %s", cfg.Format)
	}
	return Logger{
		Logger: logr.New(newLogSink(h)),
		Config: cfg,
	}, nil
}

func Discard() Logger { return Logger{Logger: logr.Discard()} }

// WithValues returns a new Logger instance with additional key/value pairs.
// See Info for documentation on how key/value pairs work.
func (l Logger) WithValues(keysAndValues ...any) Logger {
	return Logger{
		Logger: l.Logger.WithValues(keysAndValues...),
		Config: l.Config,
	}
}

func (l Logger) Info(msg string, keysAndValues ...any) {
	l.Logger.Info(msg, keysAndValues...)
}

func (l Logger) Error(err error, msg string, keysAndValues ...any) {
	l.Logger.Error(err, msg, keysAndValues...)
}

func (l Logger) V(level int) Logger {
	return Logger{
		Logger: l.Logger.V(level),
		Config: l.Config,
	}
}

// toSlogLevel converts a logr v-level to a slog level.
func toSlogLevel(verbosity int) slog.Level {
	if verbosity <= 0 {
		return slog.LevelInfo
	}
	return slog.Level(-4 - (verbosity - 1))
}
