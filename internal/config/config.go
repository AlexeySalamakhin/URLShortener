package config

import (
	"encoding/json"
	"flag"
	"os"
	"strings"

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

// getConfigPathFromArgs ищет путь к конфигу в аргументах командной строки
func getConfigPathFromArgs() string {
	for i, arg := range os.Args {
		if arg == "-c" || arg == "--config" {
			if i+1 < len(os.Args) {
				return os.Args[i+1]
			}
		}
		if after, ok := strings.CutPrefix(arg, "-c="); ok {
			return after
		}
		if after, ok := strings.CutPrefix(arg, "--config="); ok {
			return after
		}
	}
	return os.Getenv("CONFIG")
}

func NewConfigs() *Config {
	var c Config

	configPath := getConfigPathFromArgs()
	if configPath != "" {
		_ = c.loadFromJSON(configPath)
	}
	_ = env.Parse(&c)

	// Регистрируем флаги с дефолтами из структуры
	flag.StringVar(&c.ServerAddr, "a", c.ServerAddr, "Server address")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Base URL")
	flag.StringVar(&c.File, "f", c.File, "File")
	flag.StringVar(&c.ConnectionString, "d", c.ConnectionString, "Connection string")
	flag.BoolVar(&c.EnableHTTPS, "s", c.EnableHTTPS, "Enable HTTPS mode")
	flag.StringVar(&configPath, "c", configPath, "Путь к JSON-файлу конфигурации")
	flag.StringVar(&configPath, "config", configPath, "Путь к JSON-файлу конфигурации (long)")

	flag.Parse()

	return &c
}

// loadFromJSON загружает конфиг из JSON-файла (с поддержкой комментариев).
func (c *Config) loadFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}
