package url

import "fmt"

func Example() {
	_ = MakeShortURLID("anything")

	// is not valid
	_ = Validate("")

	// is valid
	_ = Validate("123")

	result := FormatShortURL("http://base.com", "part")
	fmt.Print(result)

	// Output:
	// http://base.com/part
}
