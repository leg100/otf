// Package logr provides a logger that implements the logr interface
package logr

import (
	"context"
	"time"

	"log/slog"

	"github.com/go-logr/logr"
)

// NOTE: this has been borrowed from https://cs.opensource.google/go/x/exp/+/master:slog/record.go
const badKey = "!BADKEY"

type (
	// logSink implements the logr interface for slog
	logSink struct {
		h     slog.Handler
		depth int
	}
)

func newLogSink(h slog.Handler) *logSink {
	return &logSink{h: h}
}

func (ls *logSink) Init(ri logr.RuntimeInfo) {
	ls.depth = ri.CallDepth + 2
}

// Enabled tests whether this LogSink is enabled at the specified V-level.
func (ls *logSink) Enabled(level int) bool {
	return ls.h.Enabled(context.Background(), toSlogLevel(level))
}

// Info logs a non-error message at specified V-level with the given key/value pairs as context.
func (ls *logSink) Info(level int, msg string, keysAndValues ...any) {
	r := slog.Record{Time: time.Now(), Level: toSlogLevel(level), Message: msg}
	r.Add(keysAndValues...)
	_ = ls.h.Handle(context.Background(), r)
}

// Error logs an error, with the given message and key/value pairs as context.
func (ls *logSink) Error(err error, msg string, keysAndValues ...any) {
	r := slog.Record{Time: time.Now(), Level: slog.LevelError, Message: msg}
	r.Add("error", err)
	r.Add(keysAndValues...)
	_ = ls.h.Handle(context.Background(), r)
}

// WithValues returns a new LogSink with additional key/value pairs.
func (ls *logSink) WithValues(args ...any) logr.LogSink {
	// NOTE: code borrowed from https://cs.opensource.google/go/x/exp/+/10a50721:slog/logger.go;l=97
	var (
		attr  slog.Attr
		attrs []slog.Attr
	)
	for len(args) > 0 {
		attr, args = argsToAttr(args)
		attrs = append(attrs, attr)
	}
	c := ls.clone()
	c.h = ls.h.WithAttrs(attrs)
	return c
}

// WithName returns a new LogSink with the specified name appended in NameFieldName.
// Name elements are separated by NameSeparator.
func (ls *logSink) WithName(name string) logr.LogSink {
	return nil
}

func (ls *logSink) clone() *logSink {
	c := *ls
	return &c
}

// argsToAttr turns a prefix of the nonempty args slice into an Attr
// and returns the unconsumed portion of the slice.
// If args[0] is an Attr, it returns it, resolved.
// If args[0] is a string, it treats the first two elements as
// a key-value pair.
// Otherwise, it treats args[0] as a value with a missing key.
//
// NOTE: this has been borrowed from https://cs.opensource.google/go/x/exp/+/master:slog/record.go
func argsToAttr(args []any) (slog.Attr, []any) {
	switch x := args[0].(type) {
	case string:
		if len(args) == 1 {
			return slog.String(badKey, x), nil
		}
		a := slog.Any(x, args[1])
		a.Value = a.Value.Resolve()
		return a, args[2:]

	case slog.Attr:
		x.Value = x.Value.Resolve()
		return x, args[1:]

	default:
		return slog.Any(badKey, x), args[1:]
	}
}
