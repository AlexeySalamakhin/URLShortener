package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/models"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestPostURLHandlerText(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
		body         string
		userID       string
	}{
		{
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			expectedBody: "localhost",
			body:         "https://practicum.yandex.ru",
			userID:       "test-user",
		},
		{
			method:       http.MethodPost,
			expectedCode: http.StatusBadRequest,
			expectedBody: "",
			body:         "",
			userID:       "test-user",
		},
	}
	shortener := service.NewURLShortener(store.NewInMemoryStore())
	handler := handler.NewURLHandler(shortener, "localhost:8080")
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			ctx := context.WithValue(r.Context(), "user_id", tc.userID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.PostURLHandlerText(w, r)
			require.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			respBody, _ := io.ReadAll(w.Body)
			require.Containsf(t, tc.expectedBody, string(respBody), "Тело ответа не содержит ссылку")
		})
	}
}

func TestGetURLHandler(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		URL          string
		userID       string
	}{
		{
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
			URL:          "https://practicum.yandex.ru",
			userID:       "test-user",
		},
	}

	shortener := service.NewURLShortener(store.NewInMemoryStore())
	handler := handler.NewURLHandler(shortener, "localhost:8080")
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.URL))
			ctx := context.WithValue(r.Context(), "user_id", tc.userID)
			r = r.WithContext(ctx)
			w := httptest.NewRecorder()
			handler.PostURLHandlerText(w, r)
			shortURLByte, _ := io.ReadAll(w.Body)
			shortURL, _ := url.Parse("http://" + string(shortURLByte))
			r = httptest.NewRequest(tc.method, shortURL.Path, nil)
			w = httptest.NewRecorder()
			handler.GetURLHandler(w, r)
			require.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}

type MockShortener struct {
	mock.Mock
}

// StoreReady implements handler.URLShortener.
func (m *MockShortener) StoreReady() bool {
	return true
}

func (m *MockShortener) Shorten(ctx context.Context, url string, userID string) (string, bool) {
	args := m.Called(ctx, url, userID)
	return args.String(0), args.Bool(1)
}

func (m *MockShortener) GetOriginalURL(ctx context.Context, url string) (bool, string) {
	args := m.Called(ctx, url)
	return args.Bool(0), args.String(1)
}

func (m *MockShortener) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]models.UserURLsResponse), args.Error(1)
}

func (m *MockShortener) NewURLShortener() *MockShortener {
	return &MockShortener{}
}

func TestPostURLHandlerJson(t *testing.T) {
	testCases := []struct {
		name           string
		input          models.ShortenRequest
		mockShortKey   string
		expectedStatus int
		expectedResp   models.ShortenResponse
		userID         string
	}{
		{
			name:           "successful shortening",
			input:          models.ShortenRequest{URL: "https://example.com"},
			mockShortKey:   "abc123",
			expectedStatus: http.StatusCreated,
			expectedResp:   models.ShortenResponse{Result: "http://localhost:8080/abc123"},
			userID:         "test-user",
		},
		{
			name:           "empty url",
			input:          models.ShortenRequest{URL: ""},
			mockShortKey:   "",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   models.ShortenResponse{},
			userID:         "test-user",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(MockShortener)
			mockShortener.On("Shorten", mock.Anything, tt.input.URL, tt.userID).Return(tt.mockShortKey, false)

			handler := handler.NewURLHandler(mockShortener, "http://localhost:8080")
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			handler.PostURLHandlerJSON(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var resp models.ShortenResponse
				err := json.NewDecoder(rr.Body).Decode(&resp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	testCases := []struct {
		name           string
		userID         string
		mockURLs       []models.UserURLsResponse
		mockError      error
		expectedStatus int
		expectedURLs   []models.UserURLsResponse
	}{
		{
			name:   "successful get urls",
			userID: "test-user",
			mockURLs: []models.UserURLsResponse{
				{
					ShortURL:    "abc123",
					OriginalURL: "https://example.com",
				},
				{
					ShortURL:    "def456",
					OriginalURL: "https://example.org",
				},
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedURLs: []models.UserURLsResponse{
				{
					ShortURL:    "http://localhost:8080/abc123",
					OriginalURL: "https://example.com",
				},
				{
					ShortURL:    "http://localhost:8080/def456",
					OriginalURL: "https://example.org",
				},
			},
		},
		{
			name:           "no urls found",
			userID:         "test-user",
			mockURLs:       []models.UserURLsResponse{},
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
			expectedURLs:   nil,
		},
		{
			name:           "unauthorized",
			userID:         "",
			mockURLs:       nil,
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectedURLs:   nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(MockShortener)
			mockShortener.On("GetUserURLs", mock.Anything, tt.userID).Return(tt.mockURLs, tt.mockError)

			handler := handler.NewURLHandler(mockShortener, "http://localhost:8080")
			req, _ := http.NewRequest("GET", "/api/user/urls", nil)

			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), "user_id", tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.GetUserURLs(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var urls []models.UserURLsResponse
				err := json.NewDecoder(rr.Body).Decode(&urls)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURLs, urls)
			}
		})
	}
}
