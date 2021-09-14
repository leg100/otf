package ots

import "github.com/go-logr/logr"

type LogsWriter struct {
	runLogger RunLogger
	runID     string

	logr.Logger
}

func (lw *LogsWriter) Write(p []byte) (int, error) {
	if err := lw.runLogger(lw.runID, p, PutChunkOptions{}); err != nil {
		lw.Error(err, "unable to write logs")
		return 0, err
	}

	return len(p), nil
}

func (lw *LogsWriter) Close() error {
	opts := PutChunkOptions{End: true}

	if err := lw.runLogger(lw.runID, nil, opts); err != nil {
		lw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
