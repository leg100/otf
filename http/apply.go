package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func (s *Server) GetApply(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	apply, err := s.ApplyService().Get(r.Context(), vars["id"])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeResponse(w, r, apply)
}

func (s *Server) GetApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var opts otf.GetChunkOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk, err := s.ApplyService().GetChunk(r.Context(), vars["id"], opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(chunk.Marshal()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) UploadApplyLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	var opts otf.PutChunkOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk := otf.Chunk{
		Data:  buf.Bytes(),
		Start: opts.Start,
		End:   opts.End,
	}
	if err := s.ApplyService().PutChunk(r.Context(), vars["id"], chunk); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}
