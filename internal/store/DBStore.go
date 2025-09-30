package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

// PostgresStore реализует хранилище ссылок на базе PostgreSQL.
type PostgresStore struct {
	pool *pgxpool.Pool
}

// NewDBStore инициализирует подключение к БД и возвращает экземпляр PostgresStore.
func NewDBStore(connStr string) (*PostgresStore, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %v", err)
	}

	store := &PostgresStore{
		pool: pool,
	}

	if err := store.initDB(); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	return store, nil
}

func (s *PostgresStore) initDB() error {
	_, err := s.pool.Exec(context.Background(), `
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

// Ready проверяет доступность соединения с БД.
func (s *PostgresStore) Ready() bool {
	return s.pool.Ping(context.Background()) == nil
}

// Save сохраняет новую пару короткий/исходный URL.
func (s *PostgresStore) Save(ctx context.Context, originalURL, shortURL, userID string) error {
	_, err := s.pool.Exec(
		ctx,
		"INSERT INTO urls (short_url, original_url, user_id, is_deleted) VALUES ($1, $2, $3, FALSE)",
		shortURL, originalURL, userID,
	)
	return err
}

// GetOriginalURL возвращает исходный URL по короткому.
func (s *PostgresStore) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	var originalURL, userID string
	var deleted bool
	err := s.pool.QueryRow(
		ctx,
		"SELECT original_url, user_id, is_deleted FROM urls WHERE short_url = $1",
		shortURL,
	).Scan(&originalURL, &userID, &deleted)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserURLsResponse{}, false
		}
		return models.UserURLsResponse{}, false
	}

	return models.UserURLsResponse{ShortURL: shortURL, OriginalURL: originalURL, DeletedFlag: deleted}, true
}

// GetShortURL возвращает короткий URL по исходному или ошибку, если не найден.
func (s *PostgresStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string

	err := s.pool.QueryRow(
		ctx,
		"SELECT short_url FROM urls WHERE original_url = $1 AND is_deleted = FALSE",
		originalURL,
	).Scan(&shortURL)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrShortURLNotFound
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	return shortURL, nil
}

// SaveBatch сохраняет набор записей в транзакции.
func (s *PostgresStore) SaveBatch(records []models.URLRecord) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	batch := &pgx.Batch{}
	for _, record := range records {
		batch.Queue(
			"INSERT INTO urls (short_url, original_url, user_id, is_deleted) VALUES ($1, $2, $3, FALSE)",
			record.ShortURL, record.OriginalURL, record.UserID,
		)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

// GetUserURLs возвращает список ссылок пользователя.
func (s *PostgresStore) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	rows, err := s.pool.Query(
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

// DeleteUserURLs помечает ссылки пользователя как удалённые.
func (s *PostgresStore) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := s.pool.Exec(ctx, "UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_url = ANY($2)", userID, ids)
	return err
}

// GetStats возвращает количество не удалённых URL и уникальных пользователей.
func (s *PostgresStore) GetStats(ctx context.Context) (urls int, users int, err error) {
	row := s.pool.QueryRow(ctx, "SELECT COUNT(*) FROM urls WHERE is_deleted = FALSE")
	if err = row.Scan(&urls); err != nil {
		return
	}
	row = s.pool.QueryRow(ctx, "SELECT COUNT(DISTINCT user_id) FROM urls WHERE is_deleted = FALSE")
	if err = row.Scan(&users); err != nil {
		return
	}
	return
}

// Close закрывает пул соединений.
func (s *PostgresStore) Close() error {
	s.pool.Close()
	return nil
}
