package main

import (
	"context"

	"github.com/wellywell/shorturl/internal/compress"
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/logging"
	"github.com/wellywell/shorturl/internal/router"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/tasks"
)

type Storage interface {
	Put(ctx context.Context, key string, val string, user int) error
	Get(ctx context.Context, key string) (string, error)
	PutBatch(ctx context.Context, records ...storage.URLRecord) error
	CreateNewUser(ctx context.Context) (int, error)
	GetUserURLS(ctx context.Context, userID int) ([]storage.URLRecord, error)
	DeleteBatch(ctx context.Context, records ...storage.ToDelete) error
	Close() error
}

func main() {

	log, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	conf, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	var store Storage
	if conf.DatabaseDSN != "" {
		store, err = storage.NewDatabase(conf.DatabaseDSN)
	} else if conf.FileStoragePath != "" {
		store, err = storage.NewFileMemory(conf.FileStoragePath, storage.NewMemory())
	} else {
		store = storage.NewMemory()
	}

	if err != nil {
		panic(err)
	}
	defer store.Close()

	deleteQueue := make(chan storage.ToDelete)

	urls := handlers.NewURLsHandler(store, deleteQueue, *conf)

	r := router.NewRouter(*conf, urls, log, compress.RequestUngzipper{}, compress.ResponseGzipper{})

	go tasks.DeleteWorker(deleteQueue, store)

	err = r.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
