package http

import (
	"encoding/json"
	"net/http"
)

func (s *Server) CacheStats(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(s.CacheService.Stats())
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	w.Header().Set("Content-type", jsonApplication)
	w.Write(payload)
}
