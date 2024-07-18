package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
)

var mockConfig = config.ServerConfig{BaseAddress: "localhost:8080", ShortURLsAddress: "http://localhost:8080"}

var DBDSN string

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

/*func TestMain(m *testing.M) {
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

}*/

func randomString() string {
	const charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, 20)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func TestHandleCreateShortURL(t *testing.T) {

	testCases := []struct {
		method         string
		expectedCode   int
		body           io.Reader
		bodyIsExpected bool
	}{
		{method: http.MethodGet, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodPut, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodDelete, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodPost, body: strings.NewReader("http://some_url.de"), expectedCode: http.StatusCreated, bodyIsExpected: true},
		{method: http.MethodPost, body: strings.NewReader(""), expectedCode: http.StatusBadRequest, bodyIsExpected: false},
		{method: http.MethodPost, body: strings.NewReader(strings.Repeat("a", 256)), expectedCode: http.StatusBadRequest, bodyIsExpected: false},
	}

	storage := storage.NewMemory()
	urls := &URLsHandler{urls: storage, config: mockConfig}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", tc.body)
			w := httptest.NewRecorder()

			urls.HandleCreateShortURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.bodyIsExpected {
				assert.NotEmpty(t, w.Body.String(), "Body пустой")
			}
		})
	}
}


// skip database no docker on github

/*
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
}*/

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

func TestHandleShortenURLJSON(t *testing.T) {

	testCases := []struct {
		method         string
		expectedCode   int
		body           io.Reader
		bodyIsExpected bool
	}{
		{method: http.MethodGet, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodPut, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodDelete, body: nil, expectedCode: http.StatusMethodNotAllowed, bodyIsExpected: false},
		{method: http.MethodPost, body: strings.NewReader("{\"url\": \"http://some_url.de\"}"), expectedCode: http.StatusCreated, bodyIsExpected: true},
		{method: http.MethodPost, body: strings.NewReader("{\"smth\": \"w\"}"), expectedCode: http.StatusBadRequest, bodyIsExpected: false},
	}

	storage := storage.NewMemory()
	urls := &URLsHandler{urls: storage, config: mockConfig}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/api/shorten", tc.body)
			w := httptest.NewRecorder()

			urls.HandleShortenURLJSON(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if tc.bodyIsExpected {
				assert.NotEmpty(t, w.Body.String(), "Body пустой")

				var result struct {
					Result string `json:"result"`
				}
				body, _ := io.ReadAll(w.Body)
				err := json.Unmarshal(body, &result)
				assert.NoError(t, err, "Coulnd not unmarshal result")
				assert.NotEmpty(t, result.Result, "Result пустой")
			}
		})
	}
}

/*
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
}*/

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

func TestHandleGetFullURL(t *testing.T) {

	storage := storage.NewMemory()
	urls := &URLsHandler{urls: storage, config: mockConfig}

	// Create short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://something.com"))
	w := httptest.NewRecorder()

	urls.HandleCreateShortURL(w, r)
	shortURL := w.Body.String()
	splits := strings.Split(shortURL, "/")
	urlID := splits[len(splits)-1]

	testCases := []struct {
		method       string
		url          string
		urlID        string
		expectedCode int
	}{
		{method: http.MethodPost, url: shortURL, urlID: urlID, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodPut, url: shortURL, urlID: urlID, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodDelete, url: shortURL, urlID: urlID, expectedCode: http.StatusMethodNotAllowed},
		{method: http.MethodGet, url: shortURL, urlID: urlID, expectedCode: http.StatusTemporaryRedirect},
		{method: http.MethodGet, url: "http://localhost:8080/I_dont_exist", urlID: "I_dont_exist", expectedCode: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, tc.url, nil)
			r.SetPathValue("id", tc.urlID)
			w := httptest.NewRecorder()

			urls.HandleGetFullURL(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			if w.Code == http.StatusTemporaryRedirect {
				assert.Equal(t, []string([]string{"http://something.com"}), w.Header()["Location"], "Неправильная ссылка")
			}
		})
	}
}

/*
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
}*/

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

/*func BenchmarkHandleShortenBatchDB(b *testing.B) {

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
}*/

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

/*func BenchmarkHandleUserURLSDB(b *testing.B) {

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
}*/

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

/*func BenchmarkHandleDeleteUserURLSDB(b *testing.B) {

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
}*/

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
