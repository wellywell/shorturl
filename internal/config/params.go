package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	BaseAddress      string `env:"SERVER_ADDRESS"`
	ShortURLsAddress string `env:"BASE_URL"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN      string `env:"DATABASE_DSN"`
}

func NewConfig() (*ServerConfig, error) {
	var params ServerConfig
	err := env.Parse(&params)
	if err != nil {
		return nil, err
	}

	var commandLineParams ServerConfig

	flag.StringVar(&commandLineParams.BaseAddress, "a", "localhost:8080", "Base address to listen on")
	flag.StringVar(&commandLineParams.ShortURLsAddress, "b", "http://localhost:8080", "Short URLs base address")
	flag.StringVar(&commandLineParams.FileStoragePath, "f", "/tmp/short-url-db.json", "Path to file to store urls")
	flag.StringVar(&commandLineParams.DatabaseDSN, "d", "postgres://postgres:postgres@localhost:5432/postgres", "Database DSN")
	flag.Parse()

	if params.BaseAddress == "" {
		params.BaseAddress = commandLineParams.BaseAddress
	}
	if params.ShortURLsAddress == "" {
		params.ShortURLsAddress = commandLineParams.ShortURLsAddress
	}
	if params.FileStoragePath == "" {
		params.FileStoragePath = commandLineParams.FileStoragePath
	}
	if params.DatabaseDSN == "" {
		params.DatabaseDSN = commandLineParams.DatabaseDSN
	}

	return &params, nil
}
