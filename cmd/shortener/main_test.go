package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AlexeySalamakhin/URLShortener/internal/handler"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/stretchr/testify/require"
)

func TestPostURLHandler(t *testing.T) {
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
	handler := handler.NewURLHandler(shortener)
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			w := httptest.NewRecorder()
			handler.PostURLHandler(w, r)
			require.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
			respBody, _ := io.ReadAll(r.Body)
			require.Containsf(t, tc.expectedBody, string(respBody), "Тело ответа не содержит ссылку")
		})
	}
}
