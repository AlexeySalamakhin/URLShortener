package store

import (
	"context"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

type InMemoryStore struct {
	db map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{db: make(map[string]string)}
}

func (s *InMemoryStore) Save(ctx context.Context, originalURL string, shortURL string) error {
	s.db[shortURL] = string(originalURL)
	return nil
}

func (s *InMemoryStore) GetOriginalURL(ctx context.Context, shortURL string) (found bool, originalURL string) {
	originalURL, found = s.db[shortURL]
	if !found {
		return false, ""
	}
	return true, originalURL
}

func (s *InMemoryStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	for k, v := range s.db {
		if v == originalURL {
			return k, nil
		}
	}

	return "", ErrShortURLNotFound
}

func (s *InMemoryStore) Ready() bool {
	return true
}

func (s *InMemoryStore) SaveBatch(records []models.URLRecord) error {
	var err error
	for _, record := range records {
		err = s.Save(context.Background(), record.OriginalURL, record.ShortURL)
	}
	return err
}
