package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ServerConfig interface {
	GetShortURLsAddress() string
	GetBaseAddress() string
}

type UrlsHandlers interface {
	HandleCreateShortURL(w http.ResponseWriter, req *http.Request)
	HandleGetFullURL(w http.ResponseWriter, req *http.Request)
}

type Router struct {
	config ServerConfig
	router *chi.Mux
}

func NewRouter(config ServerConfig, handlers UrlsHandlers) *Router {

	r := chi.NewRouter()

	r.Post("/", handlers.HandleCreateShortURL)
	r.Get("/{id}", handlers.HandleGetFullURL)

	return &Router{router: r, config: config}
}

func (r *Router) ListenAndServe() error {
	err := http.ListenAndServe(r.config.GetBaseAddress(), r.router)
	return err
}
