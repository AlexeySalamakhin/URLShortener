package store

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
	"sync"

	"github.com/AlexeySalamakhin/URLShortener/internal/models"
)

type FileStore struct {
	mu       sync.RWMutex      // Мьютекс для защиты доступа к данным
	db       map[string]string // shortURL -> originalURL
	file     *os.File
	writer   *bufio.Writer
	nextUUID int
}

func NewFileStore(filePath string) (*FileStore, error) {
	// Открываем файл для чтения и записи (создаем если не существует)
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	store := &FileStore{
		db:     make(map[string]string),
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

func (s *FileStore) Save(originalURL, shortURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Генерируем новый UUID
	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	// Сохраняем в памяти
	s.db[shortURL] = originalURL

	// Создаем запись для сохранения
	record := models.URLRecord{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
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

func (s *FileStore) GetOriginalURL(shortURL string) (found bool, originalURL string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	originalURL, found = s.db[shortURL]
	return
}

func (s *FileStore) GetShortURL(originalURL string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.db {
		if v == originalURL {
			return k
		}
	}

	return ""
}

func (s *FileStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.writer.Flush(); err != nil {
		return err
	}
	return s.file.Close()
}

func (s *FileStore) loadFromFile() error {
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var record models.URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return err
		}
		s.db[record.ShortURL] = record.OriginalURL
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

func (s *FileStore) Ready() bool {
	return true
}

func (s *FileStore) SaveBatch(records []models.URLRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	for _, record := range records {
		err = s.Save(record.OriginalURL, record.ShortURL)
	}
	return err
}
