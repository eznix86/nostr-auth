package server

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	buildCacheControl  = "public, max-age=31536000, immutable"
	publicCacheControl = "public, max-age=3600"
)

func BuildCacheControl() string {
	return buildCacheControl
}

func PublicCacheControl() string {
	return publicCacheControl
}

func StaticHeaders(cacheControl string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cacheControl != "" {
				w.Header().Set("Cache-Control", cacheControl)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func Gzip() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !acceptsGzip(r) {
				next.ServeHTTP(w, r)
				return
			}

			gz := gzip.NewWriter(w)
			defer gz.Close()

			wrapped := &gzipResponseWriter{
				ResponseWriter: w,
				Writer:         gz,
			}
			wrapped.Header().Add("Vary", "Accept-Encoding")
			wrapped.Header().Set("Content-Encoding", "gzip")
			wrapped.Header().Del("Content-Length")

			next.ServeHTTP(wrapped, r)
		})
	}
}

type gzipResponseWriter struct {
	http.ResponseWriter
	io.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.Header().Del("Content-Length")
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *gzipResponseWriter) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	return w.Writer.Write(data)
}

func acceptsGzip(r *http.Request) bool {
	if strings.Contains(r.Header.Get("Connection"), "Upgrade") {
		return false
	}

	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}
