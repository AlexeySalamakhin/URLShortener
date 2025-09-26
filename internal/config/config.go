package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/caarlos0/env"
)

// Config содержит конфигурационные параметры приложения, доступные через
// флаги командной строки, переменные окружения и JSON-файл.
type Config struct {
	// ServerAddr — адрес, на котором запускается сервер (например, ":8080")
	ServerAddr string `env:"SERVER_ADDRESS" json:"server_address"`
	// BaseURL — базовый URL сервиса сокращения ссылок
	BaseURL string `env:"BASE_URL" json:"base_url"`
	// File — путь к файлу для хранения данных (если используется файловое хранилище)
	File string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	// ConnectionString — строка подключения к базе данных (DSN)
	ConnectionString string `env:"DATABASE_DSN" json:"database_dsn"`
	// EnableHTTPS — флаг включения HTTPS
	EnableHTTPS bool `env:"ENABLE_HTTPS" json:"enable_https"`
	// ConfigPath — путь к файлу конфигурации
	ConfigPath string `env:"CONFIG" json:"config_path"`
}

// NewConfigs создаёт структуру конфигурации, парсит флаги, переменные окружения и JSON-файл.
func NewConfigs() *Config {
	var c Config
	c.parseFlags()

	if c.ConfigPath == "" {
		c.ConfigPath = os.Getenv("CONFIG")
	}
	if c.ConfigPath != "" {
		_ = c.loadFromJSON(c.ConfigPath)
	}
	env.Parse(&c)
	return &c
}

// parseFlags настраивает и регистрирует флаги командной строки.
func (c *Config) parseFlags() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&c.File, "f", "urls.txt", "File")
	flag.StringVar(&c.ConnectionString, "d", "", "Connection string")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "Enable HTTPS mode")
	flag.StringVar(&c.ConfigPath, "c", "", "Путь к JSON-файлу конфигурации")
	flag.StringVar(&c.ConfigPath, "config", "", "Путь к JSON-файлу конфигурации (long)")
}

// loadFromJSON загружает конфиг из JSON-файла (с поддержкой комментариев).
func (c *Config) loadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}
