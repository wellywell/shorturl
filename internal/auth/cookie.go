package auth

import (
	"net/http"
)

const userCookie = "_user"

func VerifyUser(r *http.Request) (int, error) {
	cookie, err := r.Cookie(userCookie)
	if err == nil {
		userID, err := GetUserID(cookie.Value)
		if err != nil {
			return 0, err
		}
		return userID, nil
	}
	return 0, err
}

func SetAuthCookie(userID int, w http.ResponseWriter) error {

	token, err := BuildJWTString(userID)
	if err != nil {
		return err
	}
	cookie := &http.Cookie{Name: userCookie, Value: token}
	http.SetCookie(w, cookie)
	return nil
}
