package main

import (
	"fmt"
	"net/http"

	"github.com/wellywell/shorturl/internal/storage"
)

func main() {

	storage := storage.NewMemory()
	urls := &UrlsHandler{urls: storage, host: "http://localhost:8080"}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /", urls.HandleCreateShortUrl)
	mux.HandleFunc("GET /{id}", urls.HandleGetFullURL)
	mux.HandleFunc("/", BadRequest)

	fmt.Println("Starting")
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
