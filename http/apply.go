package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	 "github.com/leg100/otf/http/jsonapi"
)

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	applyID := mux.Vars(r)["apply_id"]
	runID := otf.ConvertID(applyID, "run")

	run, err := s.Application.GetRun(r.Context(), runID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	jsonapi.WriteResponse(w, r, &apply{run.Apply(), r, s})
}

type apply struct {
	*otf.Apply
	req *http.Request
	*Server
}

// ToJSONAPI assembles a JSONAPI DTO.
func (a *apply) ToJSONAPI() any {
	dto := &jsonapi.Apply{
		ID:               otf.ConvertID(a.ID(), "apply"),
		LogReadURL:       a.signedLogURL(a.req, a.ID(), "apply"),
		Status:           string(a.Status()),
		StatusTimestamps: &jsonapi.PhaseStatusTimestamps{},
	}
	if a.ResourceReport != nil {
		dto.Additions = &a.Additions
		dto.Changes = &a.Changes
		dto.Destructions = &a.Destructions
	}
	for _, ts := range a.StatusTimestamps() {
		switch ts.Status {
		case otf.PhasePending:
			dto.StatusTimestamps.PendingAt = &ts.Timestamp
		case otf.PhaseCanceled:
			dto.StatusTimestamps.CanceledAt = &ts.Timestamp
		case otf.PhaseErrored:
			dto.StatusTimestamps.ErroredAt = &ts.Timestamp
		case otf.PhaseFinished:
			dto.StatusTimestamps.FinishedAt = &ts.Timestamp
		case otf.PhaseQueued:
			dto.StatusTimestamps.QueuedAt = &ts.Timestamp
		case otf.PhaseRunning:
			dto.StatusTimestamps.StartedAt = &ts.Timestamp
		case otf.PhaseUnreachable:
			dto.StatusTimestamps.UnreachableAt = &ts.Timestamp
		}
	}
	return dto
}
