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
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func (s *PostgresStore) Ready() bool {
	return s.db.Ping() == nil
}

func (s *PostgresStore) Save(ctx context.Context, originalURL, shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2)",
		shortURL, originalURL,
	)
	return err
}

func (s *PostgresStore) GetOriginalURL(ctx context.Context, shortURL string) (found bool, originalURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	err := s.db.QueryRowContext(
		ctx,
		"SELECT original_url FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ""
		}
		return false, ""
	}

	return true, originalURL
}

func (s *PostgresStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shortURL := ""

	err := s.db.QueryRowContext(
		ctx,
		"SELECT short_url FROM urls WHERE original_url = $1",
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
		"INSERT INTO urls (short_url, original_url) VALUES ($1, $2)",
	)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, record := range records {
		_, err = stmt.ExecContext(context.Background(), record.ShortURL, record.OriginalURL)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
