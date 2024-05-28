package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wellywell/shorturl/internal/config"
)

type URLsHandlers interface {
	HandleCreateShortURL(w http.ResponseWriter, req *http.Request)
	HandleGetFullURL(w http.ResponseWriter, req *http.Request)
	HandleShortenURLJSON(w http.ResponseWriter, req *http.Request)
	HandlePing(w http.ResponseWriter, req *http.Request)
	HandleShortenBatch(w http.ResponseWriter, req *http.Request)
	HandleUserURLS(w http.ResponseWriter, req *http.Request)
	HandleDeleteUserURLS(w http.ResponseWriter, req *http.Request)
}

type Middleware interface {
	Handle(h http.Handler) http.Handler
}

type Router struct {
	config config.ServerConfig
	router *chi.Mux
}

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

func (r *Router) ListenAndServe() error {
	addr := r.config.BaseAddress
	err := http.ListenAndServe(addr, r.router)
	return err
}
