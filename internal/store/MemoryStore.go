package store

import "github.com/AlexeySalamakhin/URLShortener/internal/models"

type InMemoryStore struct {
	db map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{db: make(map[string]string)}
}

func (s *InMemoryStore) Save(originalURL string, shortURL string) error {
	s.db[shortURL] = string(originalURL)
	return nil
}

func (s *InMemoryStore) GetOriginalURL(shortURL string) (found bool, originalURL string) {
	originalURL, found = s.db[shortURL]
	if !found {
		return false, ""
	}
	return true, originalURL
}
func (s *InMemoryStore) GetShortURL(originalURL string) string {

	for k, v := range s.db {
		if v == originalURL {
			return k
		}
	}

	return ""
}

func (s *InMemoryStore) Ready() bool {
	return true
}

func (s *InMemoryStore) SaveBatch(records []models.URLRecord) error {

	var err error
	for _, rerecord := range records {
		err = s.Save(rerecord.OriginalURL, rerecord.ShortURL)
	}
	return err
}
