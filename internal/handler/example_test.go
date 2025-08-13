package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/middleware"
	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

type fakeShortener struct {
	ready            bool
	originalByShort  map[string]models.UserURLsResponse
	userURLsByUserID map[string][]models.UserURLsResponse
	keys             []string
	nextKeyIndex     int
}

func (f *fakeShortener) Shorten(ctx context.Context, originalURL string, userID string) (string, bool) {
	if f.nextKeyIndex < len(f.keys) {
		k := f.keys[f.nextKeyIndex]
		f.nextKeyIndex++
		return k, false
	}
	return "abc123", false
}

func (f *fakeShortener) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	if f.originalByShort == nil {
		return models.UserURLsResponse{}, false
	}
	rec, ok := f.originalByShort[shortURL]
	return rec, ok
}

func (f *fakeShortener) StoreReady() bool { return f.ready }

func (f *fakeShortener) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	return f.userURLsByUserID[userID], nil
}

func (f *fakeShortener) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	return nil
}

func newTestHandler(baseURL string, s *fakeShortener) *handler.URLHandler {
	return handler.NewURLHandler(s, baseURL)
}

func ExampleURLHandler_PostURLHandlerJSON() {
	s := &fakeShortener{keys: []string{"k1"}}
	h := newTestHandler("http://example.com", s)

	body := map[string]string{"url": "http://long.example"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.PostURLHandlerJSON(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))
	// Output:
	// 201
	// {"result":"http://example.com/k1"}
}

func ExampleURLHandler_PostURLHandlerText() {
	s := &fakeShortener{keys: []string{"k1"}}
	h := newTestHandler("http://example.com", s)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://long.example"))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.PostURLHandlerText(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))
	// Output:
	// 201
	// http://example.com/k1
}

func ExampleURLHandler_GetURLHandler() {
	s := &fakeShortener{originalByShort: map[string]models.UserURLsResponse{
		"abc123": {OriginalURL: "http://long.example"},
	}}
	h := newTestHandler("http://example.com", s)

	req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
	rr := httptest.NewRecorder()

	h.GetURLHandler(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(rr.Header().Get("Location"))
	// Output:
	// 307
	// http://long.example
}

func ExampleURLHandler_Batch() {
	s := &fakeShortener{keys: []string{"k1", "k2"}}
	h := newTestHandler("http://example.com", s)

	batch := []map[string]string{{"correlation_id": "1", "original_url": "http://a"}, {"correlation_id": "2", "original_url": "http://b"}}
	b, _ := json.Marshal(batch)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Batch(rr, req)

	fmt.Println(rr.Code)
	fmt.Println(strings.TrimSpace(rr.Body.String()))
	// Output:
	// 201
	// [{"correlation_id":"1","short_url":"http://example.com/k1"},{"correlation_id":"2","short_url":"http://example.com/k2"}]
}

func ExampleURLHandler_GetUserURLs() {
	s := &fakeShortener{userURLsByUserID: map[string][]models.UserURLsResponse{
		"user-1": {{ShortURL: "k1", OriginalURL: "http://a"}, {ShortURL: "k2", OriginalURL: "http://b"}},
	}}
	h := newTestHandler("http://example.com", s)

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.GetUserURLs(rr, req)

	fmt.Println(rr.Code)
	// тело содержит массив URL-ов пользователя
	fmt.Println(strings.ReplaceAll(strings.TrimSpace(rr.Body.String()), " ", ""))
	// Output:
	// 200
	// [{"short_url":"http://example.com/k1","original_url":"http://a","is_deleted":false},{"short_url":"http://example.com/k2","original_url":"http://b","is_deleted":false}]
}

func ExampleURLHandler_DeleteUserURLs() {
	s := &fakeShortener{}
	h := newTestHandler("http://example.com", s)

	b, _ := json.Marshal([]string{"k1", "k2"})
	req := httptest.NewRequest(http.MethodDelete, "/api/user/urls", bytes.NewReader(b))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "user-1"))
	rr := httptest.NewRecorder()

	h.DeleteUserURLs(rr, req)

	fmt.Println(rr.Code)
	// Output:
	// 202
}

func ExampleURLHandler_Ping() {
	s := &fakeShortener{ready: true}
	h := newTestHandler("http://example.com", s)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rr := httptest.NewRecorder()

	h.Ping(rr, req)

	fmt.Println(rr.Code)
	// Output:
	// 200
}
