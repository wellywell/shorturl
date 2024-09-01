package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wellywell/shorturl/internal/auth"
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers"
	pb "github.com/wellywell/shorturl/internal/handlers/grpc/proto"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Storage - интерфейс хранилища коротких ссылок
type Storage interface {
	Put(ctx context.Context, key string, val string, user int) error
	Get(ctx context.Context, key string) (string, error)
	PutBatch(ctx context.Context, records ...storage.URLRecord) error
	CreateNewUser(ctx context.Context) (int, error)
	GetUserURLS(ctx context.Context, userID int) ([]storage.URLRecord, error)
	CountURLs(ctx context.Context) (int, error)
	CountUsers(ctx context.Context) (int, error)
}

// ShorturlServer поддерживает все необходимые методы сервера.
type ShorturlServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedShortURLServiceServer

	urls        Storage
	deleteQueue chan storage.ToDelete
	config      config.ServerConfig
}

// NewURLsHandler инициализирует URLsHandler, необходимого для работы хендлеров
func NewShorturlServer(storage Storage, queue chan storage.ToDelete, config config.ServerConfig) *ShorturlServer {
	return &ShorturlServer{
		urls:        storage,
		deleteQueue: queue,
		config:      config,
	}
}

// ShortenURL метод для сокращения ссылки
func (s *ShorturlServer) ShortenURL(ctx context.Context, in *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	userID, err := s.getOrCreateUser(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Error authenticating or creating user")
	}

	err = s.setAuth(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Error authenticating user")
	}
	if !url.Validate(in.Url) {
		return nil, status.Errorf(codes.InvalidArgument, "URL must be of length from 1 to 250")
	}

	shortURL, isCreated, err := handlers.GetShortURL(ctx, in.Url, userID, s.urls, s.config)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not store url")
	}

	result := pb.ShortenURLResponse{
		Result:    shortURL,
		IsCreated: isCreated,
	}
	return &result, nil
}

// ShortenBatch сокращает набор ссылок
func (s *ShorturlServer) ShortenBatch(ctx context.Context, in *pb.ShortenBatchRequest) (*pb.ShortenBatchResponse, error) {
	userID, err := s.getOrCreateUser(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Error authenticating user")
	}

	err = s.setAuth(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Error authenticating user")
	}

	if len(in.Data) == 0 {
		return &pb.ShortenBatchResponse{}, nil
	}

	records := make([]storage.URLRecord, len(in.Data))
	respData := make([]*pb.ShortenBatchOutData, len(in.Data))

	for i, data := range in.Data {
		shortURLID := url.MakeShortURLID(data.OriginalUrl)

		respData[i] = &pb.ShortenBatchOutData{
			CorrelationId: data.CorrelationId,
			ShortUrl:      url.FormatShortURL(s.config.ShortURLsAddress, shortURLID),
		}
		records[i] = storage.URLRecord{
			ShortURL: shortURLID,
			FullURL:  data.OriginalUrl,
			UserID:   userID,
		}
	}
	err = s.urls.PutBatch(ctx, records...)
	if err != nil {
		// В случае возникновения коллизий тут, завершаемся с ошибкой
		return nil, status.Errorf(codes.Internal, "Could not store values")
	}
	return &pb.ShortenBatchResponse{Data: respData}, nil
}

// GetFullURL получить длинную ссылку по id короткой
func (s *ShorturlServer) GetFullURL(ctx context.Context, in *pb.FullURLRequest) (*pb.FullURLResponse, error) {
	if in.ShortId == "" {
		return nil, status.Errorf(codes.InvalidArgument, "Id not passed")
	}
	url, err := s.urls.Get(ctx, in.ShortId)

	if err != nil {
		var keyNotFound *storage.KeyNotFoundError
		if errors.As(err, &keyNotFound) {
			return nil, status.Errorf(codes.NotFound, "Not found")
		}
		var keyDeleted *storage.RecordIsDeleted
		if errors.As(err, &keyDeleted) {
			return nil, status.Errorf(codes.ResourceExhausted, "Gone")
		}
		return nil, status.Errorf(codes.Internal, "Unknown")
	}
	return &pb.FullURLResponse{FullUrl: url}, nil
}

// Ping проверка работоспособности сервиса
func (s *ShorturlServer) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	conn, err := pgx.Connect(ctx, s.config.DatabaseDSN)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "Database unaccessable")
	}
	defer func() {
		err := conn.Close(ctx)
		if err != nil {
			fmt.Println(err)
		}
	}()
	return &pb.PingResponse{}, nil
}

// DeleteUserURLS удаляет урлы по запросу
func (s *ShorturlServer) DeleteUserURLS(ctx context.Context, in *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	user, err := s.getUser(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unknown user")
	}

	for _, rec := range in.Data {
		s.deleteQueue <- storage.ToDelete{UserID: user, ShortURL: rec}
	}
	return &pb.DeleteUserURLsResponse{}, nil
}

// GetUserURLS вернёт все урлы пользователя
func (s *ShorturlServer) GetUserURLs(ctx context.Context, in *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	user, err := s.getUser(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Unknown user")
	}
	urls, err := s.urls.GetUserURLS(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unkwnown error")
	}
	if len(urls) == 0 {
		return &pb.GetUserURLsResponse{}, nil
	}
	respData := make([]*pb.URLData, len(urls))
	for i, data := range urls {

		respData[i] = &pb.URLData{
			ShortUrl:    url.FormatShortURL(s.config.ShortURLsAddress, data.ShortURL),
			OriginalUrl: data.FullURL,
		}
	}
	return &pb.GetUserURLsResponse{Data: respData}, nil
}

// GetStats возвращает статистику
func (s *ShorturlServer) GetStats(ctx context.Context, in *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	users, err := s.urls.CountUsers(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not count users")
	}
	urls, err := s.urls.CountURLs(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not count urls")
	}
	return &pb.GetStatsResponse{Urls: int32(urls), Users: int32(users)}, nil
}

func (s *ShorturlServer) getUser(ctx context.Context) (int, error) {
	var token string

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		values := md.Get("token")
		if len(values) > 0 {
			// ключ содержит слайс строк, получаем первую строку
			token = values[0]
		}
	}

	if token != "" {
		userID, err := auth.GetUserID(token)
		if err == nil {
			return userID, nil
		}
	}
	return 0, fmt.Errorf("not authorized")
}

func (s *ShorturlServer) getOrCreateUser(ctx context.Context) (int, error) {

	var userID int

	userID, err := s.getUser(ctx)

	if err == nil {
		return userID, err
	}

	// user not verified, create new one
	userID, err = s.urls.CreateNewUser(ctx)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (s *ShorturlServer) setAuth(ctx context.Context, userID int) error {
	token, err := auth.BuildJWTString(userID)
	if err != nil {
		return err
	}
	header := metadata.Pairs("token", token)
	err = grpc.SetHeader(ctx, header)
	if err != nil {
		return err
	}
	return nil
}
