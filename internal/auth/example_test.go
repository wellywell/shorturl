package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

func Example() {

	w := httptest.NewRecorder()

	_ = SetAuthCookie(123, w)

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://some.url"))
	_, _ = VerifyUser(r)

	token, _ := BuildJWTString(123)

	userID, _ := GetUserID(token)
	fmt.Print(userID)

	// Output:
	// 123

}
