package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type ResponseGzipper struct {
	writer *gzip.Writer
}

type RequestUngzipper struct {
	reader *gzip.Reader
}

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (u RequestUngzipper) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger, err := zap.NewDevelopment()
		sugar := logger.Sugar()

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			sugar.Infoln("Content-Encoding not gzip")
			return
		}
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

		logger, err := zap.NewDevelopment()
		sugar := logger.Sugar()

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			sugar.Infoln("Accept-Encoding not gzip")
			return
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") && !strings.Contains(contentType, "text/html") {
			next.ServeHTTP(w, r)
			return
		}

		if g.writer == nil {
			g.writer, err = gzip.NewWriterLevel(w, gzip.BestCompression)
		} else {
			g.writer.Reset(w)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		defer g.writer.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: g.writer}, r)
	})
}
