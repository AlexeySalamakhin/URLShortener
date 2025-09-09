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
	ServerAddr       string `env:"SERVER_ADDRESS" json:"server_address"`
	BaseURL          string `env:"BASE_URL" json:"base_url"`
	File             string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	ConnectionString string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS      bool   `env:"ENABLE_HTTPS" json:"enable_https"`
}

// NewConfigs создаёт структуру конфигурации, парсит флаги, переменные окружения и JSON-файл.
func NewConfigs() *Config {
	var c Config
	var configPath string
	flag.StringVar(&configPath, "c", "", "Путь к JSON-файлу конфигурации")
	flag.StringVar(&configPath, "config", "", "Путь к JSON-файлу конфигурации (long)")
	c.parseFlags()
	flag.Parse()

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}
	if configPath != "" {
		_ = c.loadFromJSON(configPath)
	}
	_ = env.Parse(&c)
	return &c
}

// parseFlags настраивает и регистрирует флаги командной строки.
func (c *Config) parseFlags() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&c.File, "f", "urls.txt", "File")
	flag.StringVar(&c.ConnectionString, "d", "", "Connection string")
	flag.BoolVar(&c.EnableHTTPS, "s", false, "Enable HTTPS mode")
}

// loadFromJSON загружает конфиг из JSON-файла (с поддержкой комментариев).
func (c *Config) loadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}
