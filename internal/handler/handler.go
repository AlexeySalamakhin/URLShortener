package handler

import (
	"fmt"
	"io"
	"net/http"

	Logger "github.com/AlexeySalamakhin/URLShortener/internal/logger"
	"github.com/go-chi/chi"
)

type URLShortener interface {
	Shorten(originalURL string) string
	GetOriginalURL(shortURL string) (found bool, originalURL string)
}

type URLHandler struct {
	Shortener URLShortener
	BaseURL   string
}

func NewURLHandler(shortener URLShortener, baseURL string) *URLHandler {
	return &URLHandler{Shortener: shortener, BaseURL: baseURL}
}
func (h *URLHandler) SetupRouter() *chi.Mux {
	rout := chi.NewRouter()
	rout.Use(Logger.RequestLogger)
	rout.Post("/", h.PostURLHandler)
	rout.Get("/{shortURL}", h.GetURLHandler)
	rout.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return rout
}

func (h *URLHandler) PostURLHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	originalURL, err := io.ReadAll(r.Body)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	if len(originalURL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortKey := h.Shortener.Shorten(string(originalURL))
	w.WriteHeader(201)
	w.Write(fmt.Appendf(nil, "%s/%s", h.BaseURL, shortKey))
}

func (h *URLHandler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	found, originalURL := h.Shortener.GetOriginalURL(shortURL)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
