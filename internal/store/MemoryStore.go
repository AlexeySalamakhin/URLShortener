package store

import (
	"context"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

type InMemoryStore struct {
	db map[string]models.URLRecord
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{db: make(map[string]models.URLRecord)}
}

func (s *InMemoryStore) Save(ctx context.Context, originalURL string, shortURL string, userID string) error {
	s.db[shortURL] = models.URLRecord{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}
	return nil
}

func (s *InMemoryStore) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	record, found := s.db[shortURL]
	if !found {
		return models.UserURLsResponse{}, false
	}
	return models.UserURLsResponse{ShortURL: record.ShortURL, OriginalURL: record.OriginalURL, DeletedFlag: record.DeletedFlag}, true
}

func (s *InMemoryStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	for k, v := range s.db {
		if v.OriginalURL == originalURL {
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
		err = s.Save(context.Background(), record.OriginalURL, record.ShortURL, record.UserID)
	}
	return err
}

func (s *InMemoryStore) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	var urls []models.UserURLsResponse
	for _, record := range s.db {
		if record.UserID == userID && !record.DeletedFlag {
			urls = append(urls, models.UserURLsResponse{
				ShortURL:    record.ShortURL,
				OriginalURL: record.OriginalURL,
			})
		}
	}
	return urls, nil
}

func (s *InMemoryStore) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	for _, id := range ids {
		record, ok := s.db[id]
		if ok && record.UserID == userID && !record.DeletedFlag {
			record.DeletedFlag = true
			s.db[id] = record
		}
	}
	return nil
}
