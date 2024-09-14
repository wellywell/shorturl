package router

import (
	"fmt"

	"github.com/wellywell/shorturl/internal/compress"
	"github.com/wellywell/shorturl/internal/config"
	"github.com/wellywell/shorturl/internal/handlers/http/handlers"
	"github.com/wellywell/shorturl/internal/logging"
	"github.com/wellywell/shorturl/internal/storage"
)

func Example() {

	var mockConfig = config.ServerConfig{BaseAddress: "localhost:8080", ShortURLsAddress: "http://localhost:8080"}
	st := storage.NewMemory()
	handler := handlers.NewURLsHandler(st, make(chan storage.ToDelete), mockConfig)

	logger, _ := logging.NewLogger()

	r := NewServer(mockConfig, handler, logger, compress.RequestUngzipper{}, compress.ResponseGzipper{})

	go func() {
		err := r.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()

	// Output:

}
