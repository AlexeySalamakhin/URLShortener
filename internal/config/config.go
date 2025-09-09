package config

import (
	"flag"

	"github.com/caarlos0/env"
)

// Config содержит конфигурационные параметры приложения, доступные через
// флаги командной строки и переменные окружения.
type Config struct {
	ServerAddr       string `env:"SERVER_ADDRESS"`
	BaseURL          string `env:"BASE_URL"`
	File             string `env:"FILE_STORAGE_PATH"`
	ConnectionString string `env:"DATABASE_DSN"`
	EnableHTTPS      bool   `env:"ENABLE_HTTPS"`
}

// NewConfigs создаёт структуру конфигурации, парсит флаги и переменные окружения.
func NewConfigs() *Config {
	var c Config
	c.parseFlags()
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
}
