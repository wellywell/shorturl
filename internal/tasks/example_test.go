package tasks

import "github.com/wellywell/shorturl/internal/storage"

func Example() {
	st := storage.NewMemory()

	ch := make(chan storage.ToDelete)

	go DeleteWorker(ch, st)

	// Output:
}
