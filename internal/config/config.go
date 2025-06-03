package config

import (
	"flag"

	"github.com/caarlos0/env"
)

type Config struct {
	ServerAddr       string `env:"SERVER_ADDRESS"`
	BaseURL          string `env:"BASE_URL"`
	File             string `env:"FILE_STORAGE_PATH"`
	ConnectionString string `env:"DATABASE_DSN"`
}

func NewConfigs() *Config {
	var c Config
	c.parseFlags()
	env.Parse(&c)
	return &c
}

func (c *Config) parseFlags() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8080", "Base URL")
	flag.StringVar(&c.File, "f", "urls.txt", "File")
	flag.StringVar(&c.ConnectionString, "d", "", "Connection string")
	flag.Parse()

}
