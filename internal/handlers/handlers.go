package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/url"
)

type Storage interface {
	Put(key string, val string) error
	Get(key string) (val string, ok bool)
}

type UrlsHandler struct {
	urls   Storage
	config config.ServerConfig
}

func NewUrlsHandler(storage Storage, config config.ServerConfig) *UrlsHandler {
	return &UrlsHandler{
		urls:   storage,
		config: config,
	}
}

func (uh *UrlsHandler) HandleShortenURLJSON(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
		return
	}

	var data struct {
		Url string `json:"url"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Could not parse body",
			http.StatusBadRequest)
		return
	}

	longURL := data.Url
	if !url.Validate(longURL) {
		http.Error(w, "Url must be of length from 1 to 250", // TODO more informative answer
			http.StatusBadRequest)
		return
	}

	shortURL, err := uh.createShortURL(longURL)
	if err != nil {
		http.Error(w, "Could not store url",
			http.StatusInternalServerError)
		return
	}

	result := struct {
		Result string `json:"result"`
	}{
		Result: shortURL,
	}

	response, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Could not serialize result",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
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

	shortURL, err := uh.createShortURL(longURL)
	if err != nil {
		http.Error(w, "Could not store url",
			http.StatusInternalServerError)
	}

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *UrlsHandler) createShortURL(longURL string) (string, error) {
	shortURLID := url.MakeShortURLID(longURL)

	// Handle collisions
	for {
		err := uh.urls.Put(shortURLID, longURL)
		if err == nil {
			break
		}

		var keyExists *storage.KeyExistsError
		if errors.As(err, &keyExists) {
			// сгенерить новую ссылку и попробовать заново
			shortURLID = url.MakeShortURLID(longURL)
		} else {
			return "", err
		}

	}
	return fmt.Sprintf("%s/%s", uh.config.ShortURLsAddress, shortURLID), nil
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
