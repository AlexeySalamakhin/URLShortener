package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	logger "github.com/AlexeySalamakhin/URLShortener/internal/logger"
	"github.com/AlexeySalamakhin/URLShortener/internal/middleware"
	"github.com/AlexeySalamakhin/URLShortener/internal/models"
	"github.com/go-chi/chi"
	"go.uber.org/zap"
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
	rout.Use(middleware.RequestLogger)
	rout.Use(middleware.GzipMiddleware)
	rout.Post("/", h.PostURLHandlerText)
	rout.Post("/api/shorten", h.PostURLHandlerJSON)
	rout.Get("/{shortURL}", h.GetURLHandler)
	rout.NotFound(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})
	return rout
}

func (h *URLHandler) PostURLHandlerText(w http.ResponseWriter, r *http.Request) {
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

func (h *URLHandler) PostURLHandlerJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var req models.ShortenRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("Failed to decode JSON request", zap.Error(err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if _, err := url.ParseRequestURI(req.URL); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	shortKey := h.Shortener.Shorten(string(req.URL))

	resp := models.ShortenResponse{Result: fmt.Sprintf("%s/%s", h.BaseURL, shortKey)}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		logger.Log.Error("Failed to encode JSON request", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(jsonResp)
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
