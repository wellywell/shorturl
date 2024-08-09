package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"
)

// Claims нужен для кастомизации формата JWT token
type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

const secret = "secret"

// BuildJWTString формирует jwt-токен, включающий userID
func BuildJWTString(userID int) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{},

		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID получает id пользователя из JWT-токена
func GetUserID(tokenString string) (int, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		})
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("token invalid")
	}

	return claims.UserID, nil
}
