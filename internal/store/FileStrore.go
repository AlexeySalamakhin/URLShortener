package store

import (
	"bufio"
	"encoding/json"
	"os"
	"strconv"
)

type URLRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStore struct {
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

	// Генерируем новый UUID
	s.nextUUID++
	uuid := strconv.Itoa(s.nextUUID)

	// Сохраняем в памяти
	s.db[shortURL] = originalURL

	// Создаем запись для сохранения
	record := URLRecord{
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

func (s *FileStore) Get(shortURL string) (found bool, originalURL string) {

	originalURL, found = s.db[shortURL]
	return
}

func (s *FileStore) Close() error {
	if err := s.writer.Flush(); err != nil {
		return err
	}
	return s.file.Close()
}

func (s *FileStore) loadFromFile() error {
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var record URLRecord
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

	// _, err := s.file.Seek(0, 2)
	return nil
}
