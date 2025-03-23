package store

type URLStore interface {
	Save(originalUrl string, shortUrl string)
	Get(shortUrl string) (found bool, originalURL string)
}
