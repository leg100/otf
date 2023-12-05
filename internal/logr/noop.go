package logr

import "github.com/go-logr/logr"

var _ logr.LogSink = (*NoopSink)(nil)

// NoopSink is a log sink that does not log anything.
type NoopSink struct{}

func (n *NoopSink) Init(info logr.RuntimeInfo)        {}
func (n *NoopSink) Enabled(level int) bool            { return false }
func (n *NoopSink) Info(int, string, ...any)          {}
func (n *NoopSink) Error(error, string, ...any)       {}
func (n *NoopSink) WithValues(...any) logr.LogSink    { return n }
func (n *NoopSink) WithName(name string) logr.LogSink { return n }

func NewNoopLogger() Logger {
	return logr.New(&NoopSink{})
}
