package store

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

// FileStore реализует файловое хранилище ссылок в формате JSONL.
type FileStore struct {
	mu       sync.RWMutex
	db       map[string]models.URLRecord
	file     *os.File
	writer   *bufio.Writer
	nextUUID int
}

// NewFileStore открывает/создаёт файл и загружает существующие записи.
func NewFileStore(filePath string) (*FileStore, error) {
	// Открываем файл для чтения и записи (создаем если не существует)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	store := &FileStore{
		db:     make(map[string]models.URLRecord),
		file:   file,
		writer: bufio.NewWriter(file),
	}

	// Загружаем существующие данные из файла
	if err := store.loadFromFile(); err != nil {
		file.Close()
		return nil, err
	}

	return store, nil
}

// Save сохраняет новую запись в памяти и файле.
func (s *FileStore) Save(ctx context.Context, originalURL, shortURL, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Генерируем новый UUID
	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	// Сохраняем в памяти
	s.db[shortURL] = models.URLRecord{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}

	// Создаем запись для сохранения
	record := models.URLRecord{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		UserID:      userID,
		DeletedFlag: false,
	}

	// Кодируем в JSON
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	// Записываем в файл
	if _, err := s.writer.Write(data); err != nil {
		return err
	}
	if _, err := s.writer.WriteString("\n"); err != nil {
		return err
	}

	// Сбрасываем буфер на диск
	return s.writer.Flush()
}

// GetOriginalURL возвращает исходный URL по короткому ключу.
func (s *FileStore) GetOriginalURL(ctx context.Context, shortURL string) (models.UserURLsResponse, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, found := s.db[shortURL]
	if !found {
		return models.UserURLsResponse{}, false
	}
	return models.UserURLsResponse{ShortURL: record.ShortURL, OriginalURL: record.OriginalURL, DeletedFlag: record.DeletedFlag}, true
}

// GetShortURL возвращает короткий ключ по исходному URL или ошибку, если не найден.
func (s *FileStore) GetShortURL(ctx context.Context, originalURL string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.db {
		if v.OriginalURL == originalURL && !v.DeletedFlag {
			return k, nil
		}
	}

	return "", ErrShortURLNotFound
}

// Close закрывает файловые ресурсы хранилища.
func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writer.Flush(); err != nil {
		return err
	}
	return s.file.Close()
}

// loadFromFile загружает записи из файла при старте.
func (s *FileStore) loadFromFile() error {
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var record models.URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}
		s.db[record.ShortURL] = record
		id, err := strconv.Atoi(record.UUID)
		if err == nil {
			s.nextUUID = id
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Ready сообщает о готовности файлового хранилища.
func (s *FileStore) Ready() bool {
	return true
}

// SaveBatch сохраняет набор записей в файл.
func (s *FileStore) SaveBatch(records []models.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	for _, record := range records {
		err = s.Save(context.Background(), record.OriginalURL, record.ShortURL, record.UserID)
	}
	return err
}

// GetUserURLs возвращает ссылки пользователя.
func (s *FileStore) GetUserURLs(ctx context.Context, userID string) ([]models.UserURLsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var urls []models.UserURLsResponse
	for _, record := range s.db {
		if record.UserID == userID && !record.DeletedFlag {
			urls = append(urls, models.UserURLsResponse{
				ShortURL:    record.ShortURL,
				OriginalURL: record.OriginalURL,
			})
		}
	}
	return urls, nil
}

// DeleteUserURLs помечает ссылки пользователя как удалённые и перезаписывает файл.
func (s *FileStore) DeleteUserURLs(ctx context.Context, userID string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed := false
	for _, id := range ids {
		record, ok := s.db[id]
		if ok && record.UserID == userID && !record.DeletedFlag {
			record.DeletedFlag = true
			s.db[id] = record
			changed = true
		}
	}
	if changed {
		s.saveAllToFile()
	}
	return nil
}

// saveAllToFile перезаписывает весь файл актуальным состоянием БД.
func (s *FileStore) saveAllToFile() {
	s.file.Truncate(0)
	s.file.Seek(0, 0)
	s.writer.Reset(s.file)
	for _, record := range s.db {
		data, _ := json.Marshal(record)
		s.writer.Write(data)
		s.writer.WriteString("\n")
	}
	s.writer.Flush()
}

// GetStats возвращает количество не удалённых URL и уникальных пользователей.
func (s *FileStore) GetStats(ctx context.Context) (urls int, users int, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userSet := make(map[string]struct{})
	for _, record := range s.db {
		if !record.DeletedFlag {
			urls++
			userSet[record.UserID] = struct{}{}
		}
	}
	users = len(userSet)
	return
}
