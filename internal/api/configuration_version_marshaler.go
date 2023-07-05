package api

import (
	"fmt"
	"time"

	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/configversion"
)

func (m *jsonapiMarshaler) toConfigurationVersion(from *configversion.ConfigurationVersion) (*types.ConfigurationVersion, error) {
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", from.ID)
	uploadURL, err := m.Sign(uploadURL, time.Hour)
	if err != nil {
		return nil, err
	}
	to := &types.ConfigurationVersion{
		ID:               from.ID,
		AutoQueueRuns:    from.AutoQueueRuns,
		Speculative:      from.Speculative,
		Source:           string(from.Source),
		Status:           string(from.Status),
		StatusTimestamps: &types.CVStatusTimestamps{},
		UploadURL:        uploadURL,
	}
	for _, ts := range from.StatusTimestamps {
		switch ts.Status {
		case configversion.ConfigurationPending:
			to.StatusTimestamps.QueuedAt = &ts.Timestamp
		case configversion.ConfigurationErrored:
			to.StatusTimestamps.FinishedAt = &ts.Timestamp
		case configversion.ConfigurationUploaded:
			to.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return to, nil
}
