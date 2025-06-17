package service

import (
	"context"
	"errors"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
	"github.com/AlexeySalamakhin/URLShortener/internal/store"
	"github.com/AlexeySalamakhin/URLShortener/internal/utils"
)

type Store interface {
	Save(ctx context.Context, originalURL string, shortURL string, userID string) error
	GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool)
	GetShortURL(ctx context.Context, originalURL string) (string, error)
	Ready() bool
	GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error)
	DeleteUserURLs(ctx context.Context, userID string, ids []string) error
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

func (u *URLShortener) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	record, found := u.store.GetOriginalURL(ctx, shortURL)
	return record, found
}

func (u *URLShortener) StoreReady() bool {
	return u.store.Ready()
}

func (u *URLShortener) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	return u.store.GetUserURLs(ctx, userID)
}

func (u *URLShortener) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	return u.store.DeleteUserURLs(ctx, userID, ids)
}
