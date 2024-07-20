package storage

import (
	"context"
	"fmt"
)

func ExampleFileMemory() {

	memory := NewMemory()
	f, _ := NewFileMemory("/tmp/a", memory)

	ctx := context.Background()
	_ = memory.Put(ctx, "key", "value", 1)

	val, _ := f.Get(ctx, "key")
	fmt.Println(val)

	_ = f.PutBatch(ctx, URLRecord{"key2", "long", 1, false}, URLRecord{"key3", "long2", 1, false})
	val, _ = f.Get(ctx, "key2")
	fmt.Println(val)

	userID, _ := f.CreateNewUser(ctx)
	fmt.Println(userID)

	urls, _ := f.GetUserURLS(ctx, 2)
	fmt.Println(len(urls))

	urls, _ = f.GetUserURLS(ctx, 1)
	fmt.Println(len(urls))

	_ = f.DeleteBatch(ctx, ToDelete{"key3", 1})
	val, _ = f.Get(ctx, "key3")
	fmt.Println(val)

	_ = f.Close()

	// Output:
	// value
	// long
	// 2
	// 0
	// 3

}
