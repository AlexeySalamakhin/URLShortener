package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const secretKey = "secret_key"

func GenerateCookie(userID string) *http.Cookie {
	if userID == "" {
		userID = uuid.New().String() // Генерация уникального ID пользователя, если не передан
	}
	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	signature := hex.EncodeToString(h.Sum(nil))

	return &http.Cookie{
		Name:    "user_id",
		Value:   userID + "|" + signature,
		Expires: time.Now().Add(24 * time.Hour),
		Path:    "/",
	}
}

func ValidateCookie(cookie *http.Cookie) (string, bool) {
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return "", false
	}

	userID := parts[0]
	signature := parts[1]

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return userID, signature == expectedSignature
}
