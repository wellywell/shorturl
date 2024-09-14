package handlers

import (
	"context"
	"errors"

	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/storage"
	"github.com/wellywell/shorturl/internal/url"
)

// Storage - интерфейс хранилища коротких ссылок
type Storage interface {
	Put(ctx context.Context, key string, val string, user int) error
}

// GetShortURL создаёт, сохраняет и возвращает короткую ссылку
func GetShortURL(ctx context.Context, longURL string, user int, st Storage, conf config.ServerConfig) (URL string, isCreated bool, err error) {
	shortURLID := url.MakeShortURLID(longURL)

	// Handle collisions
	for {
		err := st.Put(ctx, shortURLID, longURL, user)
		if err == nil {
			break
		}

		var keyExists *storage.KeyExistsError
		var valueExists *storage.ValueExistsError
		if errors.As(err, &keyExists) {
			// сгенерить новую ссылку и попробовать заново
			shortURLID = url.MakeShortURLID(longURL)
		} else if errors.As(err, &valueExists) {
			return url.FormatShortURL(conf.ShortURLsAddress, valueExists.ExistingKey), false, nil
		} else {
			return "", false, err
		}
	}
	return url.FormatShortURL(conf.ShortURLsAddress, shortURLID), true, nil
}
