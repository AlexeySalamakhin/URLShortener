package service

import (
	"context"
	"errors"

	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type Store interface {
	Save(ctx context.Context, originalURL string, shortURL string, userID string) error
	GetOriginalURL(ctx context.Context, shortURL string) (found bool, originalURL string)
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	Ready() bool
}

type URLShortener struct {
	store Store
}

func NewURLShortener(store Store) *URLShortener {
	return &URLShortener{store: store}
}

func (u *URLShortener) Shorten(ctx context.Context, originalURL string, userID string) (string, bool) {
	foundURL, err := u.store.GetShortURL(ctx, originalURL)
	if err != nil && errors.Is(err, store.ErrShortURLNotFound) {
		shortKey := utils.GenerateShortURL()
		u.store.Save(ctx, originalURL, shortKey, userID)
		return shortKey, false
	}
	return foundURL, true
}

func (u *URLShortener) GetOriginalURL(ctx context.Context, shortURL string) (bool, string) {
	return u.store.GetOriginalURL(ctx, shortURL)
}

func (u *URLShortener) StoreReady() bool {
	return u.store.Ready()
}
