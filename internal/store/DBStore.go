package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

type PostgresStore struct {
	mu       sync.RWMutex
	db       *sql.DB
	nextUUID int
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

	if err := store.loadMaxUUID(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to load max UUID: %v", err)
	}

	return store, nil
}

func (s *PostgresStore) initDB() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			uuid VARCHAR(255) PRIMARY KEY,
			short_url VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT UNIQUE NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}

func (s *PostgresStore) loadMaxUUID() error {
	var maxUUID sql.NullString
	err := s.db.QueryRow("SELECT MAX(uuid) FROM urls").Scan(&maxUUID)
	if err != nil {
		return err
	}

	if maxUUID.Valid {
		id, err := strconv.Atoi(maxUUID.String)
		if err != nil {
			return err
		}
		s.nextUUID = id
	}
	return nil
}

func (s *PostgresStore) Ready() bool {
	return s.db.Ping() == nil
}

func (s *PostgresStore) Save(originalURL, shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	_, err := s.db.ExecContext(
		context.Background(),
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		uuid, shortURL, originalURL,
	)
	return err
}

func (s *PostgresStore) GetOriginalURL(shortURL string) (found bool, originalURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	err := s.db.QueryRowContext(
		context.Background(),
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

func (s *PostgresStore) GetShortURL(shortURL string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	originalURL := ""

	err := s.db.QueryRowContext(
		context.Background(),
		"SELECT short_url FROM urls WHERE original_url = $1",
		shortURL,
	).Scan(&originalURL)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ""
		}
		return ""
	}

	return originalURL
}

func (s *PostgresStore) SaveBatch(records []models.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	defer tx.Rollback()
	if err != nil {
		return err
	}
	for _, record := range records {
		err = s.Save(record.OriginalURL, record.ShortURL)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
