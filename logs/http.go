package logs

import (
	"bytes"
	"io"
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (s *Server) getLogs(w http.ResponseWriter, r *http.Request) {
	opts := otf.GetChunkOptions{}
	if err := decode.Route(&opts, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk, err := s.GetChunk(r.Context(), opts)
	// ignore not found errors because terraform-cli may call this endpoint
	// before any logs have been written and it'll exit with an error if a not
	// found error is received.
	if err != nil && err != otf.ErrResourceNotFound {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(chunk.Data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) putLogs(w http.ResponseWriter, r *http.Request) {
	chunk := otf.Chunk{}
	if err := decode.All(&chunk, r); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk.Data = buf.Bytes()
	if err := s.PutChunk(r.Context(), chunk); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}
