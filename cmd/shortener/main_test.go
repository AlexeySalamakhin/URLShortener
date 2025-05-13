package main

import (
	"bytes"
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
	}{
		{
			method:       http.MethodPost,
			expectedCode: http.StatusCreated,
			expectedBody: "localhost",
			body:         "https://practicum.yandex.ru",
		},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", body: ""},
	}
	shortener := service.NewURLShortener(store.NewInMemoryStore())
	handler := handler.NewURLHandler(shortener, "localhost:8080")
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			handler.PostURLHandlerText(w, r)
			require.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			respBody, _ := io.ReadAll(r.Body)
			require.Containsf(t, tc.expectedBody, string(respBody), "Тело ответа не содержит ссылку")
		})
	}
}

func TestGetURLHandler(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		URL          string
	}{
		{
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
			URL:          "https://practicum.yandex.ru",
		},
	}

	shortener := service.NewURLShortener(store.NewInMemoryStore())
	handler := handler.NewURLHandler(shortener, "localhost:8080")
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.URL))
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

func (m *MockShortener) Shorten(url string) string {
	args := m.Called(url)
	return args.String(0)
}

func (m *MockShortener) GetOriginalURL(url string) (bool, string) {
	return true, url
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
	}{
		{
			name:           "successful shortening",
			input:          models.ShortenRequest{URL: "https://example.com"},
			mockShortKey:   "abc123",
			expectedStatus: http.StatusCreated,
			expectedResp:   models.ShortenResponse{Result: "http://localhost:8080/abc123"},
		},
		{
			name:           "empty url",
			input:          models.ShortenRequest{URL: ""},
			mockShortKey:   "",
			expectedStatus: http.StatusBadRequest,
			expectedResp:   models.ShortenResponse{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockShortener := new(MockShortener)
			mockShortener.On("Shorten", tt.input.URL).Return(tt.mockShortKey)

			handler := handler.NewURLHandler(mockShortener, "http://localhost:8080")
			body, _ := json.Marshal(tt.input)
			req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
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
