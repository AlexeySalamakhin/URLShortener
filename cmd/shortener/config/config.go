package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseURL    string
}

func NewConfigs() *Config {
	return &Config{}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&c.BaseURL, "b", "http://localhost:8001", "Base URL")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.ServerAddr = envRunAddr
	}
	if baseURLAddr := os.Getenv("BASE_URL"); baseURLAddr != "" {
		c.BaseURL = baseURLAddr
	}
}
