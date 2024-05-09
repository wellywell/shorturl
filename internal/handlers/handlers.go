package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/url"
)

type Storage interface {
	Put(ctx context.Context, key string, val string) error
	Get(ctx context.Context, key string) (string, error)
	PutBatch(ctx context.Context, records ...storage.KeyValue) error
}

type URLsHandler struct {
	urls   Storage
	config config.ServerConfig
}

func NewURLsHandler(storage Storage, config config.ServerConfig) *URLsHandler {
	return &URLsHandler{
		urls:   storage,
		config: config,
	}
}

func (uh *URLsHandler) HandleShortenURLJSON(w http.ResponseWriter, req *http.Request) {
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
		URL string `json:"url"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Could not parse body",
			http.StatusBadRequest)
		return
	}

	longURL := data.URL
	if !url.Validate(longURL) {
		http.Error(w, "URL must be of length from 1 to 250", // TODO more informative answer
			http.StatusBadRequest)
		return
	}

	shortURL, isCreated, err := uh.getShortURL(req.Context(), longURL)
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

	if isCreated {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *URLsHandler) HandleShortenBatch(w http.ResponseWriter, req *http.Request) {

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

	type inData struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	var requestData []inData

	err = json.Unmarshal(body, &requestData)
	if err != nil {
		http.Error(w, "Could not parse body",
			http.StatusBadRequest)
		return
	}

	type outData struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}
	respData := make([]outData, len(requestData))

	w.Header().Set("content-type", "application/json")

	if len(requestData) > 0 {

		records := make([]storage.KeyValue, len(requestData))

		for i, data := range requestData {
			shortURLID := url.MakeShortURLID(data.OriginalURL)

			respData[i] = outData{
				CorrelationID: data.CorrelationID,
				ShortURL:      url.FormatShortURL(uh.config.ShortURLsAddress, shortURLID),
			}
			records[i] = storage.KeyValue{
				Key:   shortURLID,
				Value: data.OriginalURL,
			}
		}
		err := uh.urls.PutBatch(req.Context(), records...)
		if err != nil {
			// В случае возникновения коллизий тут, завершаемся с ошибкой
			http.Error(w, "Could not store values",
				http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
	response, err := json.Marshal(respData)
	if err != nil {
		http.Error(w, "Could not serialize result",
			http.StatusInternalServerError)
		return
	}
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *URLsHandler) HandleCreateShortURL(w http.ResponseWriter, req *http.Request) {

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

	longURL := string(body)
	if !url.Validate(longURL) {
		http.Error(w, "URL must be of length from 1 to 250", // TODO more informative answer
			http.StatusBadRequest)
		return
	}

	shortURL, isCreated, err := uh.getShortURL(req.Context(), longURL)
	if err != nil {
		http.Error(w, "Could not store url",
			http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "text/plain")
	if isCreated {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	_, err = w.Write([]byte(shortURL))
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *URLsHandler) getShortURL(ctx context.Context, longURL string) (URL string, isCreated bool, err error) {
	shortURLID := url.MakeShortURLID(longURL)

	// Handle collisions
	for {
		err := uh.urls.Put(ctx, shortURLID, longURL)
		if err == nil {
			break
		}

		var keyExists *storage.KeyExistsError
		var valueExists *storage.ValueExistsError
		if errors.As(err, &keyExists) {
			// сгенерить новую ссылку и попробовать заново
			shortURLID = url.MakeShortURLID(longURL)
		} else if errors.As(err, &valueExists) {
			return url.FormatShortURL(uh.config.ShortURLsAddress, valueExists.ExistingKey), false, nil
		} else {
			return "", false, err
		}
	}
	return url.FormatShortURL(uh.config.ShortURLsAddress, shortURLID), true, nil
}

func (uh *URLsHandler) HandleGetFullURL(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}

	idString := req.PathValue("id")

	if idString == "" {
		http.Error(w, "Id not passed", http.StatusBadRequest)
		return
	}
	url, err := uh.urls.Get(req.Context(), idString)

	var keyNotFound *storage.KeyNotFoundError
	if err != nil && errors.As(err, &keyNotFound) {
		http.Error(w, "Not found", http.StatusNotFound)
		return

	}
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (uh *URLsHandler) HandlePing(w http.ResponseWriter, req *http.Request) {

	conn, err := pgx.Connect(req.Context(), uh.config.DatabaseDSN)
	if err != nil {
		http.Error(w, "Database unaccessable",
			http.StatusInternalServerError)
		return
	}
	defer conn.Close(req.Context())
	w.WriteHeader(http.StatusOK)
}
