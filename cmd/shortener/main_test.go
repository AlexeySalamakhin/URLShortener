package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPostURLHandler(t *testing.T) {
	testCases := []struct {
		method       string
		expectedCode int
		expectedBody string
		body         string
	}{
		{method: http.MethodPost, expectedCode: http.StatusCreated, expectedBody: "localhost", body: "https://practicum.yandex.ru"},
		{method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", body: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/", strings.NewReader(tc.body))
			w := httptest.NewRecorder()

			PostURLHandler(w, r)
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
		shortURL     string
		urlMapping   map[string]string
	}{
		{
			method:       http.MethodGet,
			expectedCode: http.StatusTemporaryRedirect,
			shortURL:     "ffff",
			urlMapping:   map[string]string{"ffff": "https://practicum.yandex.ru"},
		},
		{
			method:       http.MethodGet,
			expectedCode: http.StatusBadRequest,
			shortURL:     "aaa",
			urlMapping:   map[string]string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, fmt.Sprintf("/%s", tc.shortURL), nil)
			w := httptest.NewRecorder()

			GetURLHandler(r, w, tc.urlMapping)
			fmt.Print(r.Body)
			require.Equal(t, tc.expectedCode, w.Code, "Код ответа не совпадает с ожидаемым")
		})
	}
}
