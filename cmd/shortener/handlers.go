package main

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

type UrlsHandler struct {
	urls Storage
	host string
}

func (uh *UrlsHandler) HandleCreateShortURL(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}

	longURL := string(body)
	if !url.ValidateURL(longURL) {
		// TODO more informative answer
		w.WriteHeader(http.StatusBadRequest)
		return
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not store url"))
		return
	}
	var sb strings.Builder
	sb.WriteString(uh.host)
	sb.WriteString("/")
	sb.WriteString(shortURLID)

	shortURL := sb.String()

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (uh *UrlsHandler) HandleGetFullURL(w http.ResponseWriter, req *http.Request) {
	idString := req.PathValue("id")

	if idString == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	url, ok := uh.urls.Get(idString)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func BadRequest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
}
