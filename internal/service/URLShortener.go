package service

import (
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type URLShortener struct {
	store store.URLStore
}

func NewURLShortener(store store.URLStore) *URLShortener {
	return &URLShortener{store: store}
}

func (u *URLShortener) Shorten(originalURL string) string {
	shortKey := utils.GenerateShortUrl()
	u.store.Save(originalURL, shortKey)
	return shortKey
}

func (u *URLShortener) RedirectUrl(shortURL string) (bool, string) {
	return u.store.Get(shortURL)
}
