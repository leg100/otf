package http

import (
	"bytes"
	"io"
	"net/http"

	"github.com/leg100/otf"
	"github.com/leg100/otf/http/decode"
)

func getLogs(w http.ResponseWriter, r *http.Request, svc otf.LogService, id string) {
	var opts otf.GetChunkOptions
	if err := decode.Query(&opts, r.URL.Query()); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	chunk, err := svc.GetChunk(r.Context(), id, opts)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if _, err := w.Write(chunk.Marshal()); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func uploadLogs(w http.ResponseWriter, r *http.Request, svc otf.LogService, id string) {
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
	if err := svc.PutChunk(r.Context(), id, chunk); err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
}
