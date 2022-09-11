package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/leg100/zerologr"
	"github.com/mattn/go-isatty"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"
)

const DefaultLogLevel = "info"

type LoggerConfig struct {
	level string

	// Toggle log colors. Must be one of auto, true, or false.
	color string
}

// NewLoggerConfigFromFlags adds flags to the given flagset, and, after the
// flagset is parsed by the caller, the flags populate the returned logger
// config.
func NewLoggerConfigFromFlags(flags *pflag.FlagSet) *LoggerConfig {
	cfg := LoggerConfig{}

	flags.StringVarP(&cfg.level, "log-level", "l", DefaultLogLevel, "Logging level")
	flags.StringVar(&cfg.color, "log-color", "auto", "Toggle log colors: auto, true or false. Auto enables colors if using a TTY.")

	return &cfg
}

func NewLogger(cfg *LoggerConfig) (logr.Logger, error) {
	zlvl, err := zerolog.ParseLevel(cfg.level)
	if err != nil {
		return logr.Logger{}, err
	}

	// Setup logger
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}
	zerolog.DurationFieldInteger = true

	switch cfg.color {
	case "auto":
		// Disable color if stdout is not a tty
		if !isatty.IsTerminal(os.Stdout.Fd()) {
			consoleWriter.NoColor = true
		}
	case "true":
		consoleWriter.NoColor = false
	case "false":
		consoleWriter.NoColor = true
	default:
		return logr.Logger{}, fmt.Errorf("invalid choice for log color: %s. Must be one of auto, true, or false", cfg.color)
	}

	logger := zerolog.New(consoleWriter).Level(zlvl).With().Timestamp().Logger()

	if logger.GetLevel() < zerolog.InfoLevel {
		// Inform the user that logging lower than INFO threshold has been
		// enabled
		logger.WithLevel(logger.GetLevel()).Msg("custom log level enabled")
	}

	// wrap within logr wrapper
	return zerologr.NewLogger(&logger), nil
}
