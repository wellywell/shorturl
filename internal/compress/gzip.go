package compress

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type ResponseGzipper struct {
	writer *gzip.Writer
}

type RequestUngzipper struct {
	reader *gzip.Reader
}

type gzipWriter struct {
	http.ResponseWriter
	compressor *ResponseGzipper
}

func (w gzipWriter) shouldCompress() bool {
	contentType := w.Header().Get("Content-Type")
	return strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/html")

}

func (w gzipWriter) WriteHeader(status int) {
	if w.shouldCompress() {
		w.Header().Set("Content-Encoding", "gzip")

	}
	w.ResponseWriter.WriteHeader(status)
}

func (w gzipWriter) Write(b []byte) (int, error) {

	if !w.shouldCompress() {
		return w.ResponseWriter.Write(b)
	}

	var err error
	if w.compressor.writer == nil {
		w.compressor.writer, err = gzip.NewWriterLevel(w.ResponseWriter, gzip.BestCompression)
	} else {
		w.compressor.writer.Reset(w.ResponseWriter)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 0, err
	}

	defer w.compressor.writer.Close()
	return w.compressor.writer.Write(b)
}

func (u RequestUngzipper) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		var err error
		if u.reader == nil {
			u.reader, err = gzip.NewReader(r.Body)
		} else {
			err = u.reader.Reset(r.Body)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = u.reader
		defer u.reader.Close()
		next.ServeHTTP(w, r)
	})
}

func (g ResponseGzipper) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(gzipWriter{ResponseWriter: w, compressor: &g}, r)
	})
}
