package handlers

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wellywell/shorturl/internal/config"
	pb "github.com/wellywell/shorturl/internal/handlers/grpc/proto"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/tasks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var mockConfig = config.ServerConfig{BaseAddress: "localhost:8080", ShortURLsAddress: "http://localhost:8080"}

type mockServerTransportStream struct {
	Header metadata.MD
}

func (m *mockServerTransportStream) Method() string {
	return "foo"
}

func (m *mockServerTransportStream) SetHeader(md metadata.MD) error {
	m.Header = md
	return nil
}

func (m *mockServerTransportStream) SendHeader(md metadata.MD) error {
	return nil
}

func (m *mockServerTransportStream) SetTrailer(md metadata.MD) error {
	return nil
}

func TestShorturlServer_ShortenURL(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *pb.ShortenURLRequest
	}

	storage := storage.NewMemory()
	s := &ShorturlServer{
		urls:   storage,
		config: mockConfig,
	}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &mockServerTransportStream{})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"empty", args{ctx, &pb.ShortenURLRequest{}}, true},
		{"notempty", args{ctx, &pb.ShortenURLRequest{Url: "1"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.ShortenURL(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.ShortenURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.NotEmpty(t, got.Result)
			}
		})
	}
}

func TestShorturlServer_ShortenBatch(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *pb.ShortenBatchRequest
	}

	storage := storage.NewMemory()
	s := &ShorturlServer{
		urls:   storage,
		config: mockConfig,
	}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &mockServerTransportStream{})

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"simple", args{
			ctx: ctx,
			in: &pb.ShortenBatchRequest{
				Data: []*pb.ShortenBatchInData{
					{CorrelationId: "1", OriginalUrl: "2"}, {CorrelationId: "2", OriginalUrl: "3"},
				},
			},
		}, false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := s.ShortenBatch(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.ShortenBatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				assert.Equal(t, len(tt.args.in.Data), len(got.Data))
			}
		})
	}
}

func TestShorturlServer_GetFullURL(t *testing.T) {

	storage := storage.NewMemory()
	s := &ShorturlServer{
		urls:   storage,
		config: mockConfig,
	}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &mockServerTransportStream{})

	type args struct {
		ctx context.Context
		in  *pb.FullURLRequest
	}

	short, err := s.ShortenURL(ctx, &pb.ShortenURLRequest{Url: "111"})
	assert.NoError(t, err)
	shortURL := strings.Split(short.Result, "/")[3]

	tests := []struct {
		name    string
		args    args
		want    *pb.FullURLResponse
		wantErr bool
	}{
		{"not found", args{ctx, &pb.FullURLRequest{ShortId: "000"}}, nil, true},
		{"found", args{ctx, &pb.FullURLRequest{ShortId: shortURL}}, &pb.FullURLResponse{FullUrl: "111"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetFullURL(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.GetFullURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShorturlServer.GetFullURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShorturlServer_Ping(t *testing.T) {

	storage := storage.NewMemory()
	s := &ShorturlServer{
		urls:   storage,
		config: mockConfig,
	}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), &mockServerTransportStream{})

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"no db", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.Ping(ctx, &pb.PingRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.Ping() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestShorturlServer_DeleteUserURLS(t *testing.T) {
	st := storage.NewMemory()

	deleteQueue := make(chan storage.ToDelete)
	s := &ShorturlServer{
		urls:        st,
		config:      mockConfig,
		deleteQueue: deleteQueue,
	}

	mockStream := &mockServerTransportStream{}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), mockStream)

	short, err := s.ShortenURL(ctx, &pb.ShortenURLRequest{Url: "111"})
	assert.NoError(t, err)
	shortURL := strings.Split(short.Result, "/")[3]

	token := mockStream.Header.Get("token")[0]
	md := metadata.New(map[string]string{"token": token})
	tokenCtx := metadata.NewIncomingContext(ctx, md)

	type args struct {
		ctx context.Context
		in  *pb.DeleteUserURLsRequest
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"unauthorized", args{ctx, &pb.DeleteUserURLsRequest{Data: []string{shortURL}}}, true},
		{"success", args{tokenCtx, &pb.DeleteUserURLsRequest{Data: []string{shortURL}}}, false},
	}
	go tasks.DeleteWorker(deleteQueue, st)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := s.DeleteUserURLS(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.DeleteUserURLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestShorturlServer_GetUserURLs(t *testing.T) {
	st := storage.NewMemory()

	s := &ShorturlServer{
		urls:   st,
		config: mockConfig,
	}

	mockStream := &mockServerTransportStream{}

	ctx := grpc.NewContextWithServerTransportStream(context.Background(), mockStream)

	short, err := s.ShortenURL(ctx, &pb.ShortenURLRequest{Url: "111"})
	assert.NoError(t, err)
	shortURL := short.Result

	token := mockStream.Header.Get("token")[0]
	md := metadata.New(map[string]string{"token": token})
	tokenCtx := metadata.NewIncomingContext(ctx, md)

	type args struct {
		ctx context.Context
		in  *pb.GetUserURLsRequest
	}
	tests := []struct {
		name    string
		args    args
		want    *pb.GetUserURLsResponse
		wantErr bool
	}{
		{"unauthorized", args{ctx, &pb.GetUserURLsRequest{}}, nil, true},
		{"success", args{tokenCtx, &pb.GetUserURLsRequest{}}, &pb.GetUserURLsResponse{Data: []*pb.URLData{{ShortUrl: shortURL, OriginalUrl: "111"}}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetUserURLs(tt.args.ctx, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.DeleteUserURLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShorturlServer.GetStats() = %v, want %v", got, tt.want)
			}

		})
	}
}

func TestShorturlServer_GetStats(t *testing.T) {
	st := storage.NewMemory()

	s := &ShorturlServer{
		urls:   st,
		config: mockConfig,
	}

	mockStream := &mockServerTransportStream{}
	ctx := grpc.NewContextWithServerTransportStream(context.Background(), mockStream)
	_, err := s.ShortenURL(ctx, &pb.ShortenURLRequest{Url: "111"})
	assert.NoError(t, err)

	tests := []struct {
		name    string
		want    *pb.GetStatsResponse
		wantErr bool
	}{
		{"success", &pb.GetStatsResponse{Users: 1, Urls: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetStats(ctx, &pb.GetStatsRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("ShorturlServer.GetStats() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShorturlServer.GetStats() = %v, want %v", got, tt.want)
			}
		})
	}
}
