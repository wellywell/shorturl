// GRPC версия cервиса shorturl
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	_ "net/http/pprof"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers/grpc/handlers"
	pb "github.com/wellywell/shorturl/internal/handlers/grpc/proto"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/tasks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	// CountURLs количество сохраненных записей
	CountURLs(ctx context.Context) (int, error)
	// CountUsers количество сохраненных пользователей
	CountUsers(ctx context.Context) (int, error)
}

func main() {

	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s", buildVersion, buildDate, buildCommit)

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
	go tasks.DeleteWorker(deleteQueue, store)

	urls := handlers.NewShorturlServer(store, deleteQueue, *conf)

	// определяем порт для сервера
	listen, err := net.Listen("tcp", conf.BaseAddress)
	if err != nil {
		log.Fatal(err)
	}
	var s *grpc.Server
	if conf.EnableHTTPS {
		fmt.Println("Starting TLS grpc")
		creds, err := credentials.NewServerTLSFromFile("server.rsa.crt", "server.rsa.key")
		if err != nil {
			log.Fatal(err)
		}
		s = grpc.NewServer(grpc.Creds(creds))
	} else {
		s = grpc.NewServer()
	}

	// регистрируем сервис
	pb.RegisterShortURLServiceServer(s, urls)

	fmt.Println("Сервер gRPC начал работу")
	// получаем запрос gRPC
	if err := s.Serve(listen); err != nil {
		log.Fatal(err)
	}
}
