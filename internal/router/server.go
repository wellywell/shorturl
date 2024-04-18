package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wellywell/shorturl/internal/config"
)

type UrlsHandlers interface {
	HandleCreateShortURL(w http.ResponseWriter, req *http.Request)
	HandleGetFullURL(w http.ResponseWriter, req *http.Request)
}

type Middleware interface {
	Handle(h http.Handler) http.Handler
}

type Router struct {
	config config.ServerConfig
	router *chi.Mux
}

func NewRouter(config config.ServerConfig, handlers UrlsHandlers, middlewares ...Middleware) *Router {

	r := chi.NewRouter()

	for _, m := range middlewares {
		r.Use(m.Handle)
	}

	r.Post("/", handlers.HandleCreateShortURL)
	r.Get("/{id}", handlers.HandleGetFullURL)

	return &Router{router: r, config: config}
}

func (r *Router) ListenAndServe() error {
	addr := r.config.BaseAddress
	err := http.ListenAndServe(addr, r.router)
	return err
}
