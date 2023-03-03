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

func (m jsonapiMarshaler) toMarshalable(cv *otf.ConfigurationVersion) marshalable {
	uploadURL := fmt.Sprintf("/configuration-versions/%s/upload", cv.id)
	uploadURL, err := m.Sign(uploadURL, time.Hour)
	if err != nil {
		// upstream middleware converts panics to HTTP500's
		panic("signing url: " + uploadURL + "; error: " + err.Error())
	}
	return marshalable{cv, uploadURL}
}

func (m jsonapiMarshaler) toMarshableList(list *otf.ConfigurationVersionList) marshalableList {
	var items []marshalable
	for _, i := range list.Items {
		items = append(items, m.toMarshalable(i))
	}
	return marshalableList{items: items, pagination: list.Pagination}
}

type marshalable struct {
	*otf.ConfigurationVersion
	uploadURL string
}

func (m marshalable) ToJSONAPI() any {
	cv := &jsonapi.ConfigurationVersion{
		ID:               m.ID,
		AutoQueueRuns:    m.AutoQueueRuns(),
		Speculative:      m.Speculative(),
		Source:           string(m.Source()),
		Status:           string(m.Status()),
		StatusTimestamps: &jsonapi.CVStatusTimestamps{},
		UploadURL:        m.uploadURL,
	}
	for _, ts := range m.StatusTimestamps() {
		switch ts.Status {
		case ConfigurationPending:
			cv.StatusTimestamps.QueuedAt = &ts.Timestamp
		case ConfigurationErrored:
			cv.StatusTimestamps.FinishedAt = &ts.Timestamp
		case ConfigurationUploaded:
			cv.StatusTimestamps.StartedAt = &ts.Timestamp
		}
	}
	return cv
}

type marshalableList struct {
	pagination *otf.Pagination
	items      []marshalable
}

func (m marshalableList) ToJSONAPI() any {
	list := &jsonapi.ConfigurationVersionList{Pagination: m.pagination.ToJSONAPI()}
	for _, item := range m.items {
		list.Items = append(list.Items, item.ToJSONAPI().(*jsonapi.ConfigurationVersion))
	}
	return list
}
