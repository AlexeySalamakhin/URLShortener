package store

import (
	"context"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

// InMemoryStore хранит данные в памяти процесса.
type InMemoryStore struct {
	db map[string]models.URLRecord
}

// NewInMemoryStore создаёт новое in-memory хранилище.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{db: make(map[string]models.URLRecord)}
}

// Save сохраняет пару короткий/исходный URL в памяти.
func (s *InMemoryStore) Save(ctx context.Context, originalURL string, shortURL string, userID string) error {
	s.db[shortURL] = models.URLRecord{
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}
	return nil
}

// GetOriginalURL возвращает исходный URL по короткому ключу.
func (s *InMemoryStore) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	record, found := s.db[shortURL]
	if !found {
		return models.UserURLsResponse{}, false
	}
	return models.UserURLsResponse{ShortURL: record.ShortURL, OriginalURL: record.OriginalURL, DeletedFlag: record.DeletedFlag}, true
}

// GetShortURL возвращает короткий ключ по исходному URL или ошибку, если не найден.
func (s *InMemoryStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	for k, v := range s.db {
		if v.OriginalURL == originalURL {
			return k, nil
		}
	}

	return "", ErrShortURLNotFound
}

// Ready сообщает о готовности хранилища.
func (s *InMemoryStore) Ready() bool {
	return true
}

// SaveBatch сохраняет набор записей.
func (s *InMemoryStore) SaveBatch(records []models.URLRecord) error {
	var err error
	for _, record := range records {
		err = s.Save(context.Background(), record.OriginalURL, record.ShortURL, record.UserID)
	}
	return err
}

// GetUserURLs возвращает ссылки пользователя.
func (s *InMemoryStore) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	count := 0
	for _, record := range s.db {
		if record.UserID == userID && !record.DeletedFlag {
			count++
		}
	}

	urls := make([]models.UserURLsResponse, 0, count)
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

// DeleteUserURLs помечает как удалённые ссылки пользователя.
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
