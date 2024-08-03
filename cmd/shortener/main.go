// Точка запуска приложения shorturl, хранящего и возвращающего короткие ссылки вместо длинных

package main

import (
	"context"

	"net/http"
	_ "net/http/pprof"

	"github.com/wellywell/shorturl/internal/compress"
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers"
	"github.com/wellywell/shorturl/internal/logging"
	"github.com/wellywell/shorturl/internal/router"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/tasks"
)

// Storage - интерфейс хранилища для ссылок
// В роли хранилища может выступать база данных, структура в памяти, и структура памяти с записью в файл
type Storage interface {
	// Put метод для записи длинной ссылки в хранилище по ключу
	Put(ctx context.Context, key string, val string, user int) error
	// Get достаёт запись по ключу
	Get(ctx context.Context, key string) (string, error)
	// PutBatch позволяет сохранять несколько записей за раз
	PutBatch(ctx context.Context, records ...storage.URLRecord) error
	// CreateNewUser создаёт нового пользователя и возвращает его id
	CreateNewUser(ctx context.Context) (int, error)
	// GetUserURLS возвращает список ссылок для данного пользователя
	GetUserURLS(ctx context.Context, userID int) ([]storage.URLRecord, error)
	// DeleteBatch удаляет набор переданных ему ссылок
	DeleteBatch(ctx context.Context, records ...storage.ToDelete) error
	// Close корректно завершает работу хранилища
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

	// pprof c chi роутером ведёт себя странно, запустим отдельно
	go http.ListenAndServe(":8081", nil)

	err = r.ListenAndServe()
	if err != nil {
		panic(err)
	}

}
