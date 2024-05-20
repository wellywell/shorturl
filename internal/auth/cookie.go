package auth

import (
	"net/http"
)

const COOKIE_NAME = "_user"

func VerifyUser(r *http.Request) (int, error) {
	cookie, err := r.Cookie(COOKIE_NAME)
	if err == nil {
		userID, err := GetUserId(cookie.Value)
		if err != nil {
			return -1, err
		}
		return userID, nil
	}
	return -1, err
}

func SetAuthCookie(userID int, w http.ResponseWriter) error {

	token, err := BuildJWTString(userID)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{Name: COOKIE_NAME, Value: token}
	http.SetCookie(w, cookie)
	return nil
}
