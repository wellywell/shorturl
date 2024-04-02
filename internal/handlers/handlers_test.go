package handlers

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wellywell/shorturl/internal/storage"
)

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
	urls := &UrlsHandler{Urls: storage, Host: "http://localhost:8080"}

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

func TestHandleGetFullURL(t *testing.T) {

	storage := storage.NewMemory()
	urls := &UrlsHandler{Urls: storage, Host: ""}

	// Create short url
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://something.com"))
	w := httptest.NewRecorder()

	urls.HandleCreateShortURL(w, r)
	shortURL := w.Body.String()
	urlID := strings.Trim(shortURL, "/")

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
				assert.Equal(t, []string([]string{"http://something.com"}), w.HeaderMap["Location"], "Неправильная ссылка")
			}
		})
	}
}
