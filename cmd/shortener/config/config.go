package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerAddr string
	BaseUrl    string
}

func NewConfigs() *Config {
	return &Config{}
}

func (c *Config) ParseFlags() {
	flag.StringVar(&c.ServerAddr, "a", ":8080", "Server address")
	flag.StringVar(&c.BaseUrl, "b", "http://localhost:8001/", "Base URL")
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		c.ServerAddr = envRunAddr
	}
	if baseUrlAddr := os.Getenv("BASE_URL"); baseUrlAddr != "" {
		c.BaseUrl = baseUrlAddr
	}
}
