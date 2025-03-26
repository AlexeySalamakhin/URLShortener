package store

type InMemoryStore struct {
	db map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{db: make(map[string]string)}
}

func (s *InMemoryStore) Save(originalURL string, shortURL string) {
	s.db[shortURL] = string(originalURL)
}

func (s *InMemoryStore) Get(shortUrl string) (found bool, originalURL string) {
	originalURL, found = s.db[shortUrl]
	if !found {
		return false, ""
	}
	return true, originalURL
}
