package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/wellywell/shorturl/internal/storage"
)

func ExampleURLsHandler_HandleCreateShortURL() {
	st := storage.NewMemory()
	handler := &URLsHandler{urls: st, config: mockConfig}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://some.url"))
	handler.HandleCreateShortURL(w, r)

	sto, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), st)
	handler = &URLsHandler{urls: sto, config: mockConfig}

	handler.HandleCreateShortURL(w, r)
	// Output:

}

func ExampleURLsHandler_HandleShortenBatch() {
	st := storage.NewMemory()
	handler := &URLsHandler{urls: st, config: mockConfig}

	w := httptest.NewRecorder()
	body := "[{\"correlation_id\": \"123\", \"original_url\": \"http://smth.com\"}, {\"correlation_id\": \"1234\", \"original_url\": \"http://smth.else.com\"}]"
	r := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", strings.NewReader(body))
	handler.HandleShortenBatch(w, r)

	sto, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), st)
	handler = &URLsHandler{urls: sto, config: mockConfig}
	handler.HandleShortenBatch(w, r)
	// Output:
}

func ExampleURLsHandler_HandleUserURLS() {
	st := storage.NewMemory()
	handler := &URLsHandler{urls: st, config: mockConfig}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	handler.HandleUserURLS(w, r)

	sto, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), st)
	handler = &URLsHandler{urls: sto, config: mockConfig}

	r = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	handler.HandleUserURLS(w, r)

	// Output:

}

func ExampleURLsHandler_HandleDeleteUserURLS() {
	st := storage.NewMemory()
	handler := &URLsHandler{urls: st, config: mockConfig}

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodDelete, "/api/user/urls", nil)
	handler.HandleDeleteUserURLS(w, r)

	sto, _ := storage.NewFileMemory(fmt.Sprintf("/tmp/%s", randomString()), st)
	handler = &URLsHandler{urls: sto, config: mockConfig}

	handler.HandleDeleteUserURLS(w, r)
	// Output:

}
