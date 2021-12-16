package sql

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"
)

var _ goose.Logger = (*GooseLogger)(nil)

type GooseLogger struct {
	logr.Logger
}

func NewGooseLogger(logger logr.Logger) *GooseLogger {
	return &GooseLogger{Logger: logger}
}

func (l *GooseLogger) Fatal(v ...interface{}) {
	l.Logger.Error(nil, fmt.Sprint(v...), "component", "database")
}

func (l *GooseLogger) Fatalf(msg string, v ...interface{}) {
	l.Logger.Error(nil, fmt.Sprintf(msg, v...), "component", "database")
}

func (l *GooseLogger) Print(v ...interface{}) {
	l.Logger.Info(fmt.Sprint(v...), "component", "database")
}

func (l *GooseLogger) Println(v ...interface{}) {
	l.Logger.Info(fmt.Sprint(v...), "component", "database")
}

func (l *GooseLogger) Printf(msg string, v ...interface{}) {
	trimmed := strings.Trim(msg, "\n")
	l.Logger.Info(fmt.Sprintf(trimmed, v...), "component", "database")
}
