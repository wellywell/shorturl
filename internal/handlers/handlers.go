// Пакет handlers включает в себя реализацию хендлеров для работы сервиса
package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/wellywell/shorturl/internal/auth"
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/url"
)

// Storage - интерфейс хранилища коротких ссылок
type Storage interface {
	Put(ctx context.Context, key string, val string, user int) error
	Get(ctx context.Context, key string) (string, error)
	PutBatch(ctx context.Context, records ...storage.URLRecord) error
	CreateNewUser(ctx context.Context) (int, error)
	GetUserURLS(ctx context.Context, userID int) ([]storage.URLRecord, error)
}

// URLsHandler структура, объединяющая в себе хранилище Storage, ServerConfig и канал deleteQueue для создания тасок на удаление ссылок
type URLsHandler struct {
	urls        Storage
	config      config.ServerConfig
	deleteQueue chan storage.ToDelete
}

// NewURLsHandler инициализирует URLsHandler, необходимого для работы хендлеров
func NewURLsHandler(storage Storage, queue chan storage.ToDelete, config config.ServerConfig) *URLsHandler {
	return &URLsHandler{
		urls:        storage,
		deleteQueue: queue,
		config:      config,
	}
}

// HandleShortenURLJSON обрабатывает запрос на создание коротких ссылок в формате application/json
func (uh *URLsHandler) HandleShortenURLJSON(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(req.Body).Decode(&data)
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

	userID, err := uh.getOrCreateUser(w, req)
	if err != nil {
		http.Error(w, "Error authenticating user", http.StatusBadRequest)
		return
	}

	shortURL, isCreated, err := uh.getShortURL(req.Context(), longURL, userID)
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

// HandleShortenBatch обрабатывает пост-запрос на создание коротких ссылок батчами
func (uh *URLsHandler) HandleShortenBatch(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}

	type inData struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	var requestData []inData

	err := json.NewDecoder(req.Body).Decode(&requestData)
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
		userID, err := uh.getOrCreateUser(w, req)
		if err != nil {
			http.Error(w, "Error authenticating user", http.StatusBadRequest)
			return
		}

		records := make([]storage.URLRecord, len(requestData))

		for i, data := range requestData {
			shortURLID := url.MakeShortURLID(data.OriginalURL)

			respData[i] = outData{
				CorrelationID: data.CorrelationID,
				ShortURL:      url.FormatShortURL(uh.config.ShortURLsAddress, shortURLID),
			}
			records[i] = storage.URLRecord{
				ShortURL: shortURLID,
				FullURL:  data.OriginalURL,
				UserID:   userID,
			}
		}
		err = uh.urls.PutBatch(req.Context(), records...)
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

// HandleCreateShortURL обрабатывает запрос на создание ссылки в формате text/plain
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

	userID, err := uh.getOrCreateUser(w, req)
	if err != nil {
		http.Error(w, "Error authenticating user", http.StatusBadRequest)
		return
	}

	shortURL, isCreated, err := uh.getShortURL(req.Context(), longURL, userID)
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

func (uh *URLsHandler) getShortURL(ctx context.Context, longURL string, user int) (URL string, isCreated bool, err error) {
	shortURLID := url.MakeShortURLID(longURL)

	// Handle collisions
	for {
		err := uh.urls.Put(ctx, shortURLID, longURL, user)
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

// HandleGetFullURL обрабатывает запрос на получение длинной ссылке по id короткой
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

	if err != nil {
		var keyNotFound *storage.KeyNotFoundError
		if errors.As(err, &keyNotFound) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		var keyDeleted *storage.RecordIsDeleted
		if errors.As(err, &keyDeleted) {
			http.Error(w, "Gone", http.StatusGone)
			return
		}
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

// HandleDeleteUserURLS обрабатывает запрос на удаление ссылок, принадлежащих данному юзеру
func (uh *URLsHandler) HandleDeleteUserURLS(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.VerifyUser(req)
	if err != nil {
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		return
	}

	var requestData []string

	err = json.NewDecoder(req.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Could not parse body",
			http.StatusBadRequest)
		return
	}

	for _, rec := range requestData {
		uh.deleteQueue <- storage.ToDelete{UserID: userID, ShortURL: rec}
	}
	w.WriteHeader(http.StatusAccepted)
}

// HandleUserURLS обрабатывает запрос на получение списка ссылок, принадлежащих данному пользователю
func (uh *URLsHandler) HandleUserURLS(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Wrong method",
			http.StatusMethodNotAllowed)
		return
	}

	userID, err := auth.VerifyUser(req)
	if err != nil {
		http.Error(w, "Authorize error", http.StatusUnauthorized)
		return
	}
	urls, err := uh.urls.GetUserURLS(req.Context(), userID)

	if err != nil {
		http.Error(w, "Error getting data", http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	type outData struct {
		ShortURL    string `json:"short_url"`
		OriginalURL string `json:"original_url"`
	}
	respData := make([]outData, len(urls))

	for i, data := range urls {

		respData[i] = outData{
			ShortURL:    url.FormatShortURL(uh.config.ShortURLsAddress, data.ShortURL),
			OriginalURL: data.FullURL,
		}
	}
	response, err := json.Marshal(respData)
	if err != nil {
		http.Error(w, "Could not serialize result",
			http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Something went wrong",
			http.StatusInternalServerError)
	}
}

func (uh *URLsHandler) getOrCreateUser(w http.ResponseWriter, req *http.Request) (int, error) {

	userID, err := auth.VerifyUser(req)
	if err == nil {
		return userID, nil
	}

	// user not verified, create new one
	userID, err = uh.urls.CreateNewUser(req.Context())
	if err != nil {
		return 0, err
	}
	err = auth.SetAuthCookie(userID, w)
	if err != nil {
		return 0, err
	}
	return userID, nil
}
