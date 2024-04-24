package compress

import (
	"compress/gzip"
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

	logger, err := zap.NewDevelopment()
	_ = err
	sugar := logger.Sugar()

	if !w.shouldCompress() {
		sugar.Infoln("Content-Type not to be gzipped")
		return w.ResponseWriter.Write(b)
	}
<<<<<<< HEAD
=======
	sugar.Infoln("Hi")
>>>>>>> e40b384 (try fix)

	compressor := w.compressor

	if compressor.writer == nil {
		sugar.Infoln("Creating writer")
		compressor.writer, err = gzip.NewWriterLevel(w, gzip.BestCompression)
	} else {
		sugar.Infoln("Resetting writer")
		compressor.writer.Reset(w)
	}
	if err != nil {
		sugar.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return 0, err
	}
<<<<<<< HEAD
=======
	sugar.Infoln("Bye")

>>>>>>> e40b384 (try fix)

	defer compressor.writer.Close()
	return compressor.writer.Write(b)
}

func (u RequestUngzipper) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logger, err := zap.NewDevelopment()
		_ = err
		sugar := logger.Sugar()
		sugar.Infoln("request ungzip middleware")

		if !strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			sugar.Infoln("Content-Encoding not gzip")
			return
		}
		sugar.Infoln("Content-Encoding gzip")

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
		_ = err
		sugar := logger.Sugar()
		sugar.Infoln("Response gzip middleware")

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			sugar.Infoln("Accept-Encoding not gzip")
			return
		}
		sugar.Infoln("Accept-Encoding gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, compressor: &g}, r)
	})
}
