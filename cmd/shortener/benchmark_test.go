package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/AlexeySalamakhin/URLShortener/internal/config"
	"github.com/AlexeySalamakhin/URLShortener/internal/service"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
)

var benchConfig = config.NewConfigs()

func BenchmarkShortenLogic_InMemoryStore(b *testing.B) {
	s := service.NewURLShortenerService(store.NewInMemoryStore())
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		s.Shorten(ctx, fmt.Sprintf("https://bench/%d", i), "user1")
	}
}

func BenchmarkShortenLogic_FileStore(b *testing.B) {
	file, _ := os.CreateTemp("", "filestore_bench_*.tmp")
	defer os.Remove(file.Name())
	fs, _ := store.NewFileStore(file.Name())
	s := service.NewURLShortenerService(fs)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		s.Shorten(ctx, fmt.Sprintf("https://bench/%d", i), "user1")
	}
}

func BenchmarkShortenLogic_PostgresStore(b *testing.B) {
	cfg := benchConfig
	if cfg.ConnectionString == "" {
		b.Skip("DATABASE_DSN не задан")
	}
	dbStore, err := store.NewDBStore(cfg.ConnectionString)
	if err != nil {
		b.Fatalf("Ошибка подключения к БД: %v", err)
	}
	s := service.NewURLShortenerService(dbStore)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		s.Shorten(ctx, "https://bench/"+randomString(12), "user1")
	}
}

func BenchmarkGetUserURLsLogic_InMemoryStore(b *testing.B) {
	s := service.NewURLShortenerService(store.NewInMemoryStore())
	ctx := context.Background()
	userID := "user1"
	for i := 0; i < 1000; i++ {
		s.Shorten(ctx, fmt.Sprintf("https://bench/%d", i), userID)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetUserURLs(ctx, userID)
	}
}

func randomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
