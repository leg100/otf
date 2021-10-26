package otf

import (
	"context"

	"github.com/go-logr/logr"
)

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// JobLogsUploader uploads a chunk of logs to the server
	JobLogsUploader

	// started is used internally by the writer to determine whether the first
	// write has been prefixed with the start marker (STX).
	started bool

	// ID of job
	ID string

	logr.Logger

	ctx context.Context
}

// Write uploads a chunk of logs to the server.
func (jw *JobWriter) Write(p []byte) (int, error) {
	opts := RunUploadLogsOptions{}

	// First chunk is prefixed with STX
	if !jw.started {
		jw.started = true
		opts.Start = true
	}

	if err := jw.UploadLogs(jw.ctx, jw.ID, p, opts); err != nil {
		jw.Error(err, "unable to write logs")
		jw.started = false
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (jw *JobWriter) Close() error {
	opts := RunUploadLogsOptions{End: true}

	if err := jw.UploadLogs(jw.ctx, jw.ID, nil, opts); err != nil {
		jw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
