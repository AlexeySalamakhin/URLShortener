package middleware

import (
	"context"
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/auth"
)

type contextKey string

// UserIDKey — ключ контекста для хранения идентификатора пользователя.
const UserIDKey contextKey = "user_id"

// CookieMiddleware обеспечивает наличие валидной cookie с user_id и
// прокидывает идентификатор пользователя в контекст запроса.
func CookieMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var newCookie *http.Cookie
		userID := auth.GenerateUserID()
		cookie, err := r.Cookie("user_id")

		if err != nil {
			newCookie = auth.GenerateCookie(userID)
			http.SetCookie(w, newCookie)
		} else {
			validCoookie := auth.ValidateCookie(cookie)
			if !validCoookie {
				newCookie = auth.GenerateCookie(userID)
				http.SetCookie(w, newCookie)
			} else {
				userID = auth.GetUserID(cookie)
			}
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
