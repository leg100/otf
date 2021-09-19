package otf

import "github.com/go-logr/logr"

// JobWriter writes logs on behalf of a job.
type JobWriter struct {
	// JobLogsUploader uploads a chunk of logs to the server
	JobLogsUploader

	// ID of job
	ID string

	logr.Logger
}

// Write uploads a chunk of logs to the server.
func (jw *JobWriter) Write(p []byte) (int, error) {
	if err := jw.UploadLogs(jw.ID, p, PutChunkOptions{}); err != nil {
		jw.Error(err, "unable to write logs")
		return 0, err
	}

	return len(p), nil
}

// Close must be called to complete writing job logs
func (jw *JobWriter) Close() error {
	opts := PutChunkOptions{End: true}

	if err := jw.UploadLogs(jw.ID, nil, opts); err != nil {
		jw.Error(err, "unable to close logs")

		return err
	}
	return nil
}
