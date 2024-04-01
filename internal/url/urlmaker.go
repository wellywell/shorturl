package url

import (
	"math/rand"
	"strings"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const length = 10

func MakeShortURLID(longURL string) (shortURL string) {
	var sb strings.Builder

	for i := 0; i < length; i++ {
		letter := letters[rand.Intn(len(letters))]
		sb.WriteString(string(letter))
	}
	return sb.String()
}

func ValidateURL(url string) bool {
	return len(url) > 0 && len(url) < 250
}
