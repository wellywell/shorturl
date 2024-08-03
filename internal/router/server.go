// Пакет отвечает за инициализацию объекта роутер, матчинг набора маршрутов с нужными хендлерами-обработчикамиб
// а также подключает требуемые middleware
package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

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
}

// Middleware - интерфейс, которому должны соответствовать используемые Middleware
type Middleware interface {
	Handle(h http.Handler) http.Handler
}

// Router - объект роутера
type Router struct {
	config config.ServerConfig
	router *chi.Mux
}

// NewRouter инициализирует Router, прописывает пути, на которых сервер будет слушать
func NewRouter(config config.ServerConfig, handlers URLsHandlers, middlewares ...Middleware) *Router {

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

	return &Router{router: r, config: config}
}

// ListenAndServe - метод для запуска сервера
func (r *Router) ListenAndServe() error {
	addr := r.config.BaseAddress
	err := http.ListenAndServe(addr, r.router)
	return err
}
