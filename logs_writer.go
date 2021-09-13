package ots

import "github.com/go-logr/logr"

type LogsWriter struct {
	runLogger RunLogger
	runID     string

	// logStarted indicates whether first chunk of logs have been sent.
	logStarted bool

	logr.Logger
}

func (lw *LogsWriter) Write(p []byte) (int, error) {
	opts := AppendLogOptions{}

	if !lw.logStarted {
		opts.Start = true
	}

	if err := lw.runLogger(lw.runID, p, opts); err != nil {
		return 0, err
	}

	// Only upon success mark start chunk as having been sent.
	if !lw.logStarted {
		lw.logStarted = true
	}

	return len(p), nil
}

func (lw *LogsWriter) Close() error {
	opts := AppendLogOptions{End: true}

	if !lw.logStarted {
		opts.Start = true
	}

	if err := lw.runLogger(lw.runID, nil, opts); err != nil {
		lw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
