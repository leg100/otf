// This succint etag middleware has been borrowed from:
//
// https://github.com/wtg/shuttletracker/blob/cdd56dc4aeca922f333c913f09c1796851d6f677/api/etag.go
//
// It's very well articulated too by the author on their blog:
//
// https://sidney.kochman.org/2018/etag-middleware-go/
package http

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strings"

	"github.com/leg100/otf/internal/logr"
)

type etagResponseWriter struct {
	http.ResponseWriter
	buf  bytes.Buffer
	hash hash.Hash
	w    io.Writer
}

func (e *etagResponseWriter) Write(p []byte) (int, error) {
	return e.w.Write(p)
}

type etagMiddleware struct {
	logger logr.Logger
	// Only apply etag to paths with this prefix.
	prefix string
}

func (e *etagMiddleware) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, e.prefix) {
			next.ServeHTTP(w, r)
			return
		}
		if r.Header.Get("Accept") == "text/event-stream" {
			next.ServeHTTP(w, r)
			return
		}
		ew := &etagResponseWriter{
			ResponseWriter: w,
			buf:            bytes.Buffer{},
			hash:           sha1.New(),
		}
		ew.w = io.MultiWriter(&ew.buf, ew.hash)

		next.ServeHTTP(ew, r)

		sum := fmt.Sprintf("%x", ew.hash.Sum(nil))
		w.Header().Set("ETag", sum)

		if r.Header.Get("If-None-Match") == sum {
			// If this header isn't set then the templ proxy returns empty content
			//
			// TODO: only add this header in development
			w.Header().Set("templ-skip-modify", "true")

			w.WriteHeader(304)
		} else {
			_, err := ew.buf.WriteTo(w)
			if err != nil {
				e.logger.Error(err, "etag middleware: writing response")
			}
		}
	})
}
