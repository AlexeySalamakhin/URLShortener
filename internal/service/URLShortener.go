package service

import (
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type Store interface {
	Save(originalURL string, shortURL string) error
	Get(shortURL string) (found bool, originalURL string)
}
type URLShortener struct {
	store Store
}

func NewURLShortener(store Store) *URLShortener {
	return &URLShortener{store: store}
}

func (u *URLShortener) Shorten(originalURL string) string {
	shortKey := utils.GenerateShortURL()
	u.store.Save(originalURL, shortKey)
	return shortKey
}

func (u *URLShortener) GetOriginalURL(shortURL string) (bool, string) {
	return u.store.Get(shortURL)
}
