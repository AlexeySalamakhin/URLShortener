package handler

import (
	"context"
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
	Shorten(ctx context.Context, originalURL string, userID string) (string, bool)
	GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool)
	StoreReady() bool
	GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, userID string, ids []string)
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

	rout.Group(func(r chi.Router) {
		r.Use(middleware.CookieMiddleware)
		r.Post("/", h.PostURLHandlerText)
		r.Post("/api/shorten", h.PostURLHandlerJSON)
		r.Post("/api/shorten/batch", h.Batch)
		r.Get("/{shortURL}", h.GetURLHandler)
		r.Get("/api/user/urls", h.GetUserURLs)
		r.Delete("/api/user/urls", h.DeleteUserURLs)
	})

	rout.Get("/ping", h.Ping)
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
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	shortKey, conflict := h.Shortener.Shorten(r.Context(), string(originalURL), userID)

	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

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
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	shortKey, conflict := h.Shortener.Shorten(r.Context(), req.URL, userID)

	resp := models.ShortenResponse{Result: fmt.Sprintf("%s/%s", h.BaseURL, shortKey)}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		logger.Log.Error("Failed to encode JSON request", zap.Error(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if conflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
	w.Write(jsonResp)
}

func (h *URLHandler) GetURLHandler(w http.ResponseWriter, r *http.Request) {
	shortURL := r.URL.Path[1:]
	record, found := h.Shortener.GetOriginalURL(r.Context(), shortURL)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if record.DeletedFlag {
		w.WriteHeader(http.StatusGone)
		return
	}
	http.Redirect(w, r, record.OriginalURL, http.StatusTemporaryRedirect)
}

func (h *URLHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if h.Shortener.StoreReady() {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusInternalServerError)
}

func (h *URLHandler) Batch(w http.ResponseWriter, r *http.Request) {
	var req models.ShortURLBatchRequest
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		logger.Log.Error("Failed to decode JSON request", zap.Error(err))
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}
	var resp []models.URLBatchResponse
	for _, record := range req {
		shortURL, _ := h.Shortener.Shorten(r.Context(), record.OriginalURL, "")
		resp = append(resp, models.URLBatchResponse{
			CorrelationID: record.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.BaseURL, shortURL),
		})
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		logger.Log.Error("Failed to encode JSON request", zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonResp)
}

func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	urls, err := h.Shortener.GetUserURLs(r.Context(), userID)
	if err != nil {
		logger.Log.Error("Failed to get user URLs", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	for i := range urls {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", h.BaseURL, urls[i].ShortURL)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		logger.Log.Error("Failed to encode response", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *URLHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var ids []string
	if err := json.NewDecoder(r.Body).Decode(&ids); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	go h.Shortener.DeleteUserURLs(r.Context(), userID, ids)
	w.WriteHeader(http.StatusAccepted)
}
