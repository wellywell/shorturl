package main

import (
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/router"
	"github.com/wellywell/shorturl/internal/storage"
)

func main() {

	storage := storage.NewMemory()
	config := config.NewCommandLineParams()
	urls := handlers.NewUrlsHandler(storage, config)

	r := router.NewRouter(config, urls)

	err := r.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
