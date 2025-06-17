package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

type PostgresStore struct {
	mu sync.RWMutex
	db *sql.DB
}

func NewDBStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	store := &PostgresStore{
		db: db,
	}

	if err := store.initDB(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return store, nil
}

func (s *PostgresStore) initDB() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			uuid SERIAL PRIMARY KEY,
			short_url VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL,
			user_id VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_deleted BOOLEAN DEFAULT FALSE
		);
	`)
	return err
}

func (s *PostgresStore) Ready() bool {
	return s.db.Ping() == nil
}

func (s *PostgresStore) Save(ctx context.Context, originalURL, shortURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO urls (short_url, original_url, user_id, is_deleted) VALUES ($1, $2, $3, FALSE)",
		shortURL, originalURL, userID,
	)
	return err
}

func (s *PostgresStore) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var originalURL, userID string
	var deleted bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT original_url, user_id, is_deleted FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL, &userID, &deleted)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.UserURLsResponse{}, false
		}
		return models.UserURLsResponse{}, false
	}

	return models.UserURLsResponse{ShortURL: shortURL, OriginalURL: originalURL, DeletedFlag: deleted}, true
}

func (s *PostgresStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shortURL := ""

	err := s.db.QueryRowContext(
		ctx,
		"SELECT short_url FROM urls WHERE original_url = $1 AND is_deleted = FALSE",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrShortURLNotFound
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return shortURL, nil
}

func (s *PostgresStore) SaveBatch(records []models.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(
		context.Background(),
		"INSERT INTO urls (short_url, original_url, user_id, is_deleted) VALUES ($1, $2, $3, FALSE)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		_, err = stmt.ExecContext(context.Background(), record.ShortURL, record.OriginalURL, record.UserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.QueryContext(
		ctx,
		"SELECT short_url, original_url FROM urls WHERE user_id = $1 AND is_deleted = FALSE",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []models.UserURLsResponse
	for rows.Next() {
		var url models.UserURLsResponse
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

func (s *PostgresStore) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(ids) == 0 {
		return nil
	}
	query := "UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)"
	_, err := s.db.ExecContext(ctx, query, userID, ids)
	return err
}
