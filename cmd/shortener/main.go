package main

import (
	"fmt"
	"net/http"

	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/storage"
)

func main() {

	storage := storage.NewMemory()
	urls := &handlers.UrlsHandler{Urls: storage, Host: "http://localhost:8080"}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", urls.HandleCreateShortURL)
	mux.HandleFunc("GET /{id}", urls.HandleGetFullURL)
	mux.HandleFunc("/", handlers.BadRequest)

	fmt.Println("Starting")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
