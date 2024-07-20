package compress

import "github.com/go-chi/chi/v5"

func Example() {

	r := chi.NewRouter()
	// подключаем middleware
	unzipper := RequestUngzipper{}
	r.Use(unzipper.Handle)

	gzipper := ResponseGzipper{}
	r.Use(gzipper.Handle)

	// Output:
}
