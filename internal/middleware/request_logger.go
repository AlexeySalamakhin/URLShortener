package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/AlexeySalamakhin/URLShortener/internal/logger"
)

// RequestLogger логирует входящие HTTP-запросы и исходящие ответы с длительностью.
func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		next.ServeHTTP(&lw, r)
		duration := time.Since(start)

		logger.Log.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", duration),
		)
		logger.Log.Info("HTTP response",
			zap.Int("status code", responseData.status),
			zap.Int("size", responseData.size),
		)
	})
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

// Write записывает тело ответа и учитывает его размер для логирования.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader фиксирует статус ответа для последующего логирования.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode
}
