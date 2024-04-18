package main

import (
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/logging"
	"github.com/wellywell/shorturl/internal/router"
	"github.com/wellywell/shorturl/internal/storage"
)

func main() {

	log, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	config, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	storage := storage.NewMemory()
	urls := handlers.NewUrlsHandler(storage, *config)

	r := router.NewRouter(*config, urls, log)

	err = r.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
