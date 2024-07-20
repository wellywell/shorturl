package logging

import "github.com/go-chi/chi/v5"

func Example() {
	logger, _ := NewLogger()

	r := chi.NewRouter()
	r.Use(logger.Handle)

	// Output:

}
