// Package router отвечает за инициализацию объекта роутер, матчинг набора маршрутов с нужными хендлерами-обработчикамиб
// а также подключает требуемые middleware
package router

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wellywell/shorturl/internal/auth"
	"github.com/wellywell/shorturl/internal/config"
)

// URLsHandlers интерфейс для работы с хендлерами
type URLsHandlers interface {
	HandleCreateShortURL(w http.ResponseWriter, req *http.Request)
	HandleGetFullURL(w http.ResponseWriter, req *http.Request)
	HandleShortenURLJSON(w http.ResponseWriter, req *http.Request)
	HandlePing(w http.ResponseWriter, req *http.Request)
	HandleShortenBatch(w http.ResponseWriter, req *http.Request)
	HandleUserURLS(w http.ResponseWriter, req *http.Request)
	HandleDeleteUserURLS(w http.ResponseWriter, req *http.Request)
	HandleGetStats(w http.ResponseWriter, req *http.Request)
}

// Middleware - интерфейс, которому должны соответствовать используемые Middleware
type Middleware interface {
	Handle(h http.Handler) http.Handler
}

// Router - объект роутера
type Server struct {
	server http.Server
	config config.ServerConfig
}

// NewRouter инициализирует Router, прописывает пути, на которых сервер будет слушать
func NewServer(config config.ServerConfig, handlers URLsHandlers, middlewares ...Middleware) *Server {

	r := chi.NewRouter()

	for _, m := range middlewares {
		r.Use(m.Handle)
	}

	r.Post("/", handlers.HandleCreateShortURL)
	r.Get("/{id}", handlers.HandleGetFullURL)
	r.Post("/api/shorten", handlers.HandleShortenURLJSON)
	r.Get("/ping", handlers.HandlePing)
	r.Post("/api/shorten/batch", handlers.HandleShortenBatch)
	r.Get("/api/user/urls", handlers.HandleUserURLS)
	r.Delete("/api/user/urls", handlers.HandleDeleteUserURLS)

	r.With(auth.SubnetChecker{Trusted: config.Trusted}.Handle).Get("/api/internal/stats", handlers.HandleGetStats)

	return &Server{server: http.Server{Addr: config.BaseAddress, Handler: r}, config: config}
}

// ListenAndServe - метод для запуска сервера
func (s *Server) ListenAndServe() error {
	var err error
	if s.config.EnableHTTPS {
		err = s.server.ListenAndServeTLS("server.rsa.crt", "server.rsa.key")
	} else {
		err = s.server.ListenAndServe()

	}
	return err
}

// Shutdown gracefull shutddown
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
