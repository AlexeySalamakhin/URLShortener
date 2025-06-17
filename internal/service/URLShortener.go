package service

import (
	"context"
	"errors"
	"sync"

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

func fanIn(doneCh chan struct{}, resultChs ...chan error) chan error {
	finalCh := make(chan error)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func (u *URLShortener) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	const batchSize = 10

	doneCh := make(chan struct{})
	defer close(doneCh)

	var batches [][]string
	for i := 0; i < len(ids); i += batchSize {
		end := i + batchSize
		if end > len(ids) {
			end = len(ids)
		}
		batches = append(batches, ids[i:end])
	}

	batchChs := make([]chan error, len(batches))
	for i := range batchChs {
		batchChs[i] = make(chan error, 1)
	}

	resultCh := fanIn(doneCh, batchChs...)

	var wg sync.WaitGroup
	for i, batch := range batches {
		wg.Add(1)
		go func(batchIndex int, batch []string) {
			defer wg.Done()
			defer close(batchChs[batchIndex])

			select {
			case <-ctx.Done():
				batchChs[batchIndex] <- ctx.Err()
				return
			default:
			}

			if err := u.store.DeleteUserURLs(ctx, userID, batch); err != nil {
				batchChs[batchIndex] <- err
				return
			}
		}(i, batch)
	}

	go func() {
		wg.Wait()
	}()

	for err := range resultCh {
		if err != nil {
			return err
		}
	}

	return nil
}
