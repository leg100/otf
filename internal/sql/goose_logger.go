package sql

import (
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pressly/goose/v3"
)

var _ goose.Logger = (*gooseLogger)(nil)

type gooseLogger struct {
	logr.Logger
}

func (l *gooseLogger) Fatal(v ...interface{}) {
	l.Logger.Error(nil, fmt.Sprint(v...))
}

func (l *gooseLogger) Fatalf(msg string, v ...interface{}) {
	l.Logger.Error(nil, fmt.Sprintf(msg, v...))
}

func (l *gooseLogger) Print(v ...interface{}) {
	l.Logger.Info(fmt.Sprint(v...))
}

func (l *gooseLogger) Println(v ...interface{}) {
	l.Logger.Info(fmt.Sprint(v...))
}

func (l *gooseLogger) Printf(msg string, v ...interface{}) {
	trimmed := strings.Trim(msg, "\n")
	l.Logger.Info(fmt.Sprintf(trimmed, v...))
}
