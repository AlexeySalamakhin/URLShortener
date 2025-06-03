package service

import (
	"errors"

	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type Store interface {
	Save(originalURL string, shortURL string) error
	GetOriginalURL(shortURL string) (found bool, originalURL string)
	GetShortURL(originalURL string) (string, error)
	Ready() bool
}
type URLShortener struct {
	store Store
}

func NewURLShortener(store Store) *URLShortener {
	return &URLShortener{store: store}
}

func (u *URLShortener) Shorten(originalURL string) (string, bool) {
	foundURL, err := u.store.GetShortURL(originalURL)
	if err != nil && errors.Is(err, store.ErrShortURLNotFound) {
		shortKey := utils.GenerateShortURL()
		u.store.Save(originalURL, shortKey)
		return shortKey, false
	}
	return foundURL, true
}

func (u *URLShortener) GetOriginalURL(shortURL string) (bool, string) {
	return u.store.GetOriginalURL(shortURL)
}

func (u *URLShortener) StoreReady() bool {
	return u.store.Ready()
}
