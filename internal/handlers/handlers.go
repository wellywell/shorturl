package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/wellywell/shorturl/internal/url"
)

type Storage interface {
	Put(key string, val string) error
	Get(key string) (val string, ok bool)
}

type ServerConfig interface {
	GetShortURLsAddress() string
}

type UrlsHandler struct {
	urls   Storage
	config ServerConfig
}

func NewUrlsHandler(storage Storage, config ServerConfig) *UrlsHandler {
	return &UrlsHandler{
		urls:   storage,
		config: config,
	}
}

func (uh *UrlsHandler) HandleCreateShortURL(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}

	longURL := string(body)
	if !url.Validate(longURL) {
		http.Error(w, "Url must be of length from 1 to 250", // TODO more informative answer
			http.StatusBadRequest)
	}
	shortURLID := url.MakeShortURLID(longURL)

	// Handle collisions
	for {
		val, exists := uh.urls.Get(shortURLID)
		if exists && val != longURL {
			shortURLID = url.MakeShortURLID(longURL)
		} else {
			break
		}
	}
	err = uh.urls.Put(shortURLID, longURL)
	// TODO more specific error handling
	if err != nil {
		http.Error(w, "Could not store url",
			http.StatusInternalServerError)
	}
	var sb strings.Builder
	sb.WriteString("http:/")
	sb.WriteString(uh.config.GetShortURLsAddress())
	sb.WriteString("/")
	sb.WriteString(shortURLID)

	shortURL := sb.String()

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *UrlsHandler) HandleGetFullURL(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
	}

	idString := req.PathValue("id")

	if idString == "" {
		http.Error(w, "Id not passed", http.StatusBadRequest)
	}
	url, ok := uh.urls.Get(idString)
	if !ok {
		http.Error(w, "Not found", http.StatusNotFound)

	}
	w.Header().Set("location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func BadRequest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
