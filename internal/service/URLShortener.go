package service

import (
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type Store interface {
	Save(originalURL string, shortURL string) error
	GetOriginalURL(shortURL string) (found bool, originalURL string)
	GetShortURL(originalURL string) string
	Ready() bool
}
type URLShortener struct {
	store Store
}

func NewURLShortener(store Store) *URLShortener {
	return &URLShortener{store: store}
}

func (u *URLShortener) Shorten(originalURL string) (string, bool) {
	foundURL := u.store.GetShortURL(originalURL)
	if foundURL != "" {
		return foundURL, true
	}
	shortKey := utils.GenerateShortURL()
	u.store.Save(originalURL, shortKey)
	return shortKey, false
}

func (u *URLShortener) GetOriginalURL(shortURL string) (bool, string) {
	return u.store.GetOriginalURL(shortURL)
}

func (u *URLShortener) StoreReady() bool {
	return u.store.Ready()
}
