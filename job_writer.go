package otf

import (
	"context"

	"github.com/go-logr/logr"
	tfe "github.com/leg100/go-tfe"
)

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// JobLogsUploader uploads a chunk of logs to the server
	JobLogsUploader

	// ID of job
	ID string

	logr.Logger

	ctx context.Context
}

// Write uploads a chunk of logs to the server.
func (jw *JobWriter) Write(p []byte) (int, error) {
	if err := jw.UploadLogs(jw.ctx, jw.ID, p, tfe.RunUploadLogsOptions{}); err != nil {
		jw.Error(err, "unable to write logs")
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (jw *JobWriter) Close() error {
	opts := tfe.RunUploadLogsOptions{End: true}

	if err := jw.UploadLogs(jw.ctx, jw.ID, nil, opts); err != nil {
		jw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
