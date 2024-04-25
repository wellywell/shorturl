package main

import (
	"github.com/wellywell/shorturl/internal/compress"
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

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}
	storage, err := storage.NewFileMemory(conf.FileStoragePath, storage.NewMemory())
	if err != nil {
		panic(err)
	}
	defer storage.Close()

	urls := handlers.NewURLsHandler(storage, *conf)

	r := router.NewRouter(*conf, urls, log, compress.RequestUngzipper{}, compress.ResponseGzipper{})

	err = r.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
