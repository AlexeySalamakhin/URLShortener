package handler

import (
	"fmt"
	"io"
	"net/http"

	"github.com/AlexeySalamakhin/URLShortener/internal/service"
)

type URLHandler struct {
	Shortener *service.URLShortener
}

func NewURLHandler(shortener *service.URLShortener) *URLHandler {
	return &URLHandler{Shortener: shortener}
}

func (h *URLHandler) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		h.PostURLHandler(w, r)
	} else if r.Method == http.MethodGet {
		h.GetURLHandler(r, w)
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
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
	w.Write(fmt.Appendf(nil, "%s/%s", r.Host, shortKey))
}

func (h *URLHandler) GetURLHandler(r *http.Request, w http.ResponseWriter) {
	shortUrl := r.URL.Path[1:]
	found, originalURL := h.Shortener.RedirectUrl(shortUrl)
	if !found {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}
