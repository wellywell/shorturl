// Точка запуска приложения shorturl, хранящего и возвращающего короткие ссылки вместо длинных

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
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

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s", buildVersion, buildDate, buildCommit)

	logger, err := logging.NewLogger()
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
	defer func() {
		err = store.Close()
		if err != nil {
			panic(err)
		}
	}()

	deleteQueue := make(chan storage.ToDelete)

	urls := handlers.NewURLsHandler(store, deleteQueue, *conf)

	s := router.NewServer(*conf, urls, logger, compress.RequestUngzipper{}, compress.ResponseGzipper{})

	go tasks.DeleteWorker(deleteQueue, store)

	// pprof c chi роутером ведёт себя странно, запустим отдельно
	go func() {
		err = http.ListenAndServe(":8081", nil)
		if err != nil {
			panic(err)
		}
	}()

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	go func() {
		<-sig
		// Trigger graceful shutdown
		err := s.Shutdown(serverCtx)
		if err != nil {
			log.Fatal(err)
		}
		serverStopCtx()
	}()

	err = s.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
