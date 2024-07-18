package handlers

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/testutils"
)

var mockConfig = config.ServerConfig{BaseAddress: "localhost:8080", ShortURLsAddress: "http://localhost:8080"}

var DBDSN string

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func TestMain(m *testing.M) {
	code, err := runMain(m)

	if err != nil {
		log.Fatal(err)
	}
	os.Exit(code)
}

func runMain(m *testing.M) (int, error) {

	databaseDSN, cleanUp, err := testutils.RunTestDatabase()
	defer cleanUp()

	if err != nil {
		return 1, err
	}
	DBDSN = databaseDSN

	exitCode := m.Run()

	return exitCode, nil

}

func randomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 20)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func BenchmarkHandleCreateShortURLDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(randomString()))
		b.StartTimer()
		handler.HandleCreateShortURL(w, r)
	}
}

func BenchmarkHandleCreateShortURLMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(randomString()))
		b.StartTimer()
		handler.HandleCreateShortURL(w, r)
	}
}

func BenchmarkHandleCreateShortURLFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(randomString()))
		b.StartTimer()
		handler.HandleCreateShortURL(w, r)
	}
}

func BenchmarkHandleShortenURLJSONDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		body := strings.NewReader(fmt.Sprintf("{\"url\": \"%s\"}", randomString()))
		r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		b.StartTimer()
		handler.HandleShortenURLJSON(w, r)
	}
}

func BenchmarkHandleShortenURLJSONMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		body := strings.NewReader(fmt.Sprintf("{\"url\": \"%s\"}", randomString()))
		r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		b.StartTimer()
		handler.HandleShortenURLJSON(w, r)
	}
}

func BenchmarkHandleShortenURLJSONFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		body := strings.NewReader(fmt.Sprintf("{\"url\": \"%s\"}", randomString()))
		r := httptest.NewRequest(http.MethodPost, "/api/shorten", body)
		b.StartTimer()
		handler.HandleShortenURLJSON(w, r)
	}
}

func BenchmarkHandleGetFullURLDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	// Create short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://something.com"))
	w := httptest.NewRecorder()

	handler.HandleCreateShortURL(w, r)
	shortURL := w.Body.String()
	splits := strings.Split(shortURL, "/")
	urlID := splits[len(splits)-1]

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, shortURL, nil)
		r.SetPathValue("id", urlID)
		b.StartTimer()
		handler.HandleGetFullURL(w, r)
	}
}

func BenchmarkHandleGetFullURLMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	// Create short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://something.com"))
	w := httptest.NewRecorder()

	handler.HandleCreateShortURL(w, r)
	shortURL := w.Body.String()
	splits := strings.Split(shortURL, "/")
	urlID := splits[len(splits)-1]

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, shortURL, nil)
		r.SetPathValue("id", urlID)
		b.StartTimer()
		handler.HandleGetFullURL(w, r)
	}
}

func BenchmarkHandleGetFullURLFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	// Create short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://something.com"))
	w := httptest.NewRecorder()

	handler.HandleCreateShortURL(w, r)
	shortURL := w.Body.String()
	splits := strings.Split(shortURL, "/")
	urlID := splits[len(splits)-1]

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, shortURL, nil)
		r.SetPathValue("id", urlID)
		b.StartTimer()
		handler.HandleGetFullURL(w, r)
	}
}

func BenchmarkHandleShortenBatchDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		body := "["
		for j := 0; j <= 10; j++ {
			body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
		}
		body = body + "]"

		r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
		b.StartTimer()
		handler.HandleShortenBatch(w, r)
	}
}

func BenchmarkHandleShortenBatchMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		body := "["
		for j := 0; j <= 10; j++ {
			body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
		}
		body = body + "]"

		r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
		b.StartTimer()
		handler.HandleShortenBatch(w, r)
	}
}

func BenchmarkHandleShortenBatchFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		body := "["
		for j := 0; j <= 10; j++ {
			body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
		}
		body = body + "]"

		r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
		b.StartTimer()
		handler.HandleShortenBatch(w, r)
	}
}

func BenchmarkHandleUserURLSDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleUserURLS(w, r)
	}
}

func BenchmarkHandleUserURLSMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleUserURLS(w, r)
	}
}

func BenchmarkHandleUserURLSFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleUserURLS(w, r)
	}
}

func BenchmarkHandleDeleteUserURLSDB(b *testing.B) {

	storage, _ := storage.NewDatabase(DBDSN)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleDeleteUserURLS(w, r)
	}
}

func BenchmarkHandleDeleteUserURLSMemory(b *testing.B) {

	storage := storage.NewMemory()
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleDeleteUserURLS(w, r)
	}
}

func BenchmarkHandleDeleteUserURLSFile(b *testing.B) {

	mem := storage.NewMemory()
	storage, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), mem)
	handler := &URLsHandler{urls: storage, config: mockConfig}

	w := httptest.NewRecorder()
	body := "["
	for j := 0; j <= 10; j++ {
		body = body + fmt.Sprintf(", {\"correlation_id\": \"%s\", \"original_url\": \"%s\"}", randomString(), randomString())
	}
	body = body + "]"

	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
		b.StartTimer()
		handler.HandleDeleteUserURLS(w, r)
	}
}
