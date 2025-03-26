package store

type URLStore interface {
	Save(originalURL string, shortURL string)
	Get(shortURL string) (found bool, originalURL string)
}
