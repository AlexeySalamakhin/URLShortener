package store

import "github.com/AlexeySalamakhin/URLShortener/internal/config"

type Store interface {
	Save(originalURL string, shortURL string) error
	GetOriginalURL(shortURL string) (found bool, originalURL string)
	Ready() bool
	GetShortURL(shortURL string) string
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
