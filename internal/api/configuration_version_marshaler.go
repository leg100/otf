package api

import (
	"net/http"
	"strings"

	"github.com/DataDog/jsonapi"
	"github.com/leg100/otf/internal"
	"github.com/leg100/otf/internal/api/types"
	"github.com/leg100/otf/internal/configversion"
)

func (m *jsonapiMarshaler) toConfigurationVersion(from *configversion.ConfigurationVersion, r *http.Request) (*types.ConfigurationVersion, []jsonapi.MarshalOption) {
	to := &types.ConfigurationVersion{
		ID:               from.ID,
		AutoQueueRuns:    from.AutoQueueRuns,
		Speculative:      from.Speculative,
		Source:           string(from.Source),
		Status:           string(from.Status),
		StatusTimestamps: &types.CVStatusTimestamps{},
	}
	if from.IngressAttributes != nil {
		to.IngressAttributes = &types.IngressAttributes{
			ID: internal.ConvertID(from.ID, "ia"),
		}
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
	var included []any
	if includes := r.URL.Query().Get("include"); includes != "" {
		for _, inc := range strings.Split(includes, ",") {
			switch inc {
			case "ingress_attributes":
				if to.IngressAttributes == nil {
					break
				}
				included = append(included, &types.IngressAttributes{
					ID:        internal.ConvertID(from.ID, "ia"),
					CommitSHA: from.IngressAttributes.CommitSHA,
					CommitURL: from.IngressAttributes.CommitURL,
				})
			}
		}
	}
	opts := []jsonapi.MarshalOption{jsonapi.MarshalInclude(included...)}
	return to, opts
}
