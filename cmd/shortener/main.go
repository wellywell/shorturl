package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/storage"
)

func main() {

	storage := storage.NewMemory()
	urls := &handlers.UrlsHandler{Urls: storage, Host: "http://localhost:8080"}

	r := chi.NewRouter()

	r.Post("/", urls.HandleCreateShortURL)
	r.Get("/{id}", urls.HandleGetFullURL)

	err := http.ListenAndServe(`:8080`, r)
	if err != nil {
		panic(err)
	}
}
