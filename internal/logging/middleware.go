// Package logging реализует middleware, логгирующую информацию о сервисе
package logging

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	responseData struct {
		data   []byte
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write переопределяет метод записи ответа, добавляя к нему сохранение информации для логирования
func (r *loggingResponseWriter) Write(b []byte) (int, error) {

	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	r.responseData.data = b
	return size, err
}

// WriteHeader переопределяет метод WriteHeader для сохранения информации о статусе ответа для дальнейшего логгирования
func (r *loggingResponseWriter) WriteHeader(statusCode int) {

	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}

// Logger структура для логгирования
type Logger struct {
	sugar *zap.SugaredLogger
}

// NewLogger инициализирует логгер
func NewLogger() (*Logger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = logger.Sync()
	}()

	sugar := logger.Sugar()
	return &Logger{
		sugar: sugar,
	}, nil
}

// Handle метод для использования Logger как middleware
func (l Logger) Handle(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		l.sugar.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
	return http.HandlerFunc(logFn)
}
