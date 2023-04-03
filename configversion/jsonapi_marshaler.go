package configversion

import (
	"fmt"
	"time"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/jsonapi"
)

// jsonapiMarshaler converts config version into a struct suitable for
// marshaling into json-api
type jsonapiMarshaler struct {
	otf.Signer // for signing upload url
}

func (m *jsonapiMarshaler) toConfigurationVersion(from *ConfigurationVersion) (*jsonapi.ConfigurationVersion, error) {
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", from.ID)
	uploadURL, err := m.Sign(uploadURL, time.Hour)
	if err != nil {
		return nil, err
	}
	to := &jsonapi.ConfigurationVersion{
		ID:               from.ID,
		AutoQueueRuns:    from.AutoQueueRuns,
		Speculative:      from.Speculative,
		Source:           string(from.Source),
		Status:           string(from.Status),
		StatusTimestamps: &jsonapi.CVStatusTimestamps{},
		UploadURL:        uploadURL,
	}
	for _, ts := range from.StatusTimestamps {
		switch ts.Status {
		case ConfigurationPending:
			to.StatusTimestamps.QueuedAt = &ts.Timestamp
		case ConfigurationErrored:
			to.StatusTimestamps.FinishedAt = &ts.Timestamp
		case ConfigurationUploaded:
			to.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return to, nil
}

func (m *jsonapiMarshaler) toList(from *ConfigurationVersionList) (*jsonapi.ConfigurationVersionList, error) {
	to := &jsonapi.ConfigurationVersionList{
		Pagination: jsonapi.NewPagination(from.Pagination),
	}
	for _, i := range from.Items {
		cv, err := m.toConfigurationVersion(i)
		if err != nil {
			return nil, err
		}
		to.Items = append(to.Items, cv)
	}
	return to, nil
}
