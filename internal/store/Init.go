package store

import (
	"context"
	"errors"

	"github.com/AlexeySalamakhin/URLShortener/internal/config"
	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

var (
	ErrShortURLNotFound = errors.New("short URL not found")
)

type Store interface {
	Save(ctx context.Context, originalURL string, shortURL string, userID string) error
	GetOriginalURL(ctx context.Context, shortURL string) (found bool, originalURL string)
	Ready() bool
	GetShortURL(ctx context.Context, shortURL string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error)
}

func InitStore(cfg *config.Config) (Store, error) {
	switch {
	case cfg.ConnectionString != "":
		return NewDBStore(cfg.ConnectionString)
	case cfg.File != "":
		return NewFileStore(cfg.File)
	default:
		return NewInMemoryStore(), nil
	}
}
