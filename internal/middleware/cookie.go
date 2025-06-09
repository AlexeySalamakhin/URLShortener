package middleware

import (
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/auth"
)

func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")
		user_id, validCoookie := auth.ValidateCookie(cookie)
		if err != nil || !validCoookie {
			newCookie := auth.GenerateCookie(user_id) // Генерация уникального ID пользователя
			http.SetCookie(w, newCookie)
		}
		w.WriteHeader(http.StatusOK)
	})
}
