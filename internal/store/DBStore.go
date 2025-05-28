package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib" // PostgreSQL driver
)

type PostgresStore struct {
	mu       sync.RWMutex
	db       *sql.DB
	nextUUID int
	ready    bool
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

	store.ready = true
	return store, nil
}

func (s *PostgresStore) initDB() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			uuid VARCHAR(255) PRIMARY KEY,
			short_url VARCHAR(255) UNIQUE NOT NULL,
			original_url TEXT NOT NULL,
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

func (s *PostgresStore) Save(originalURL, shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	_, err := s.db.Exec(
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		uuid, shortURL, originalURL,
	)
	if err != nil {
		return fmt.Errorf("failed to save URL: %v", err)
	}

	return nil
}

func (s *PostgresStore) Get(shortURL string) (found bool, originalURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	err := s.db.QueryRow(
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

func (s *PostgresStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *PostgresStore) Ready() bool {
	return s.ready
}

func (s *PostgresStore) SaveWithContext(ctx context.Context, originalURL, shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	_, err := s.db.ExecContext(
		ctx,
		"INSERT INTO urls (uuid, short_url, original_url) VALUES ($1, $2, $3)",
		uuid, shortURL, originalURL,
	)
	return err
}

func (s *PostgresStore) GetWithContext(ctx context.Context, shortURL string) (found bool, originalURL string) {
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
