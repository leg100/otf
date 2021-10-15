package sqlite

import (
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

var _ goose.Logger = (*GooseLogger)(nil)

type GooseLogger struct {
	*zerolog.Logger
}

func NewGooseLogger(zlogger *zerolog.Logger) *GooseLogger {
	return &GooseLogger{Logger: zlogger}
}

func (l *GooseLogger) Fatal(v ...interface{}) {
	l.Panic().Str("component", "database").Msg(fmt.Sprint(v...))
}

func (l *GooseLogger) Fatalf(msg string, v ...interface{}) {
	l.Panic().Str("component", "database").Msg(fmt.Sprintf(msg, v...))
}

func (l *GooseLogger) Print(v ...interface{}) {
	l.Info().Str("component", "database").Msg(fmt.Sprint(v...))
}

func (l *GooseLogger) Println(v ...interface{}) {
	l.Info().Str("component", "database").Msg(fmt.Sprint(v...))
}

func (l *GooseLogger) Printf(msg string, v ...interface{}) {
	trimmed := strings.Trim(msg, "\n")
	l.Info().Str("component", "database").Msg(fmt.Sprintf(trimmed, v...))
}
