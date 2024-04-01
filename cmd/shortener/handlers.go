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

func (uh *UrlsHandler) HandleCreateShortUrl(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Something went wrong"))
		return
	}

	longUrl := string(body)
	if !url.ValidateUrl(longUrl) {
		// TODO more informative answer
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortUrlId := url.MakeShortUrlId(longUrl)

	// Handle collisions
	for {
		val, exists := uh.urls.Get(shortUrlId)
		if exists && val != longUrl {
			shortUrlId = url.MakeShortUrlId(longUrl)
		}
		break
	}
	err = uh.urls.Put(shortUrlId, longUrl)
	// TODO more specific error handling
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Could not store url"))
		return
	}
	var sb strings.Builder
	sb.WriteString(uh.host)
	sb.WriteString("/")
	sb.WriteString(shortUrlId)

	shortUrl := sb.String()

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortUrl))
}

func (uh *UrlsHandler) HandleGetFullUrl(w http.ResponseWriter, req *http.Request) {
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
