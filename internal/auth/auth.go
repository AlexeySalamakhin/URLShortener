package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const secretKey = "secret_key"

// GenerateCookie создаёт подписанную cookie с идентификатором пользователя.
func GenerateCookie(userID string) *http.Cookie {

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	signature := hex.EncodeToString(h.Sum(nil))

	return &http.Cookie{
		Name:  "user_id",
		Value: userID + "|" + signature,
	}
}

// GenerateUserID генерирует и возвращает новый уникальный идентификатор пользователя.
func GenerateUserID() string {
	return uuid.New().String()
}

// GetUserID извлекает идентификатор пользователя из значения cookie.
func GetUserID(cookie *http.Cookie) string {
	parts := strings.Split(cookie.Value, "|")
	return parts[0]
}

// ValidateCookie проверяет подпись cookie и возвращает её валидность.
func ValidateCookie(cookie *http.Cookie) bool {
	if cookie == nil {
		return false
	}
	parts := strings.Split(cookie.Value, "|")
	if len(parts) != 2 {
		return false
	}

	userID := parts[0]
	signature := parts[1]

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(userID))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return signature == expectedSignature
}
