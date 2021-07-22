package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) GetBlob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	obj, err := s.BlobService.Get(vars["id"])
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
		return
	}

	w.Write(obj)
}

func (s *Server) PutBlob(w http.ResponseWriter, r *http.Request) {
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, r.Body); err != nil {
		WriteError(w, http.StatusUnprocessableEntity, err)
		return
	}

	_, err := s.BlobService.Put(buf.Bytes())
	if err != nil {
		WriteError(w, http.StatusNotFound, err)
	}
}
