// Package config используется для конфигурирования сервиса
package config

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
)

// ServerConfig - тип для сохранения настроек сервиса
type ServerConfig struct {
	BaseAddress      string `env:"SERVER_ADDRESS" json:"server_address"`
	ShortURLsAddress string `env:"BASE_URL" json:"base_url"`
	FileStoragePath  string `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	DatabaseDSN      string `env:"DATABASE_DSN" json:"database_dsn"`
	EnableHTTPS      bool   `env:"ENABLE_HTTPS" json:"enable_https"`
	ConfigFile       string `env:"CONFIG"`
}

func parseFileParams(name string) ServerConfig {
	jsonFile, err := os.Open(name)

	if err != nil {
		log.Fatal(err)
	}
	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)

	var config ServerConfig

	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

type configValue interface {
	~bool | ~string
}

func firstNotZero[T configValue](values ...T) T {
	var zero T
	for _, v := range values {
		if v != zero {
			return v
		}
	}
	return zero
}

// NewConfig инициализация объекта ServerConfig. Параметры берутся из env, либо аргументов командной строки
func NewConfig() (*ServerConfig, error) {
	var params ServerConfig
	err := env.Parse(&params)
	if err != nil {
		return nil, err
	}

	var commandLineParams ServerConfig

	flag.StringVar(&commandLineParams.BaseAddress, "a", "", "Base address to listen on")
	flag.StringVar(&commandLineParams.ShortURLsAddress, "b", "", "Short URLs base address")
	flag.StringVar(&commandLineParams.FileStoragePath, "f", "", "Path to file to store urls")
	flag.StringVar(&commandLineParams.DatabaseDSN, "d", "", "Database DSN")
	flag.BoolVar(&commandLineParams.EnableHTTPS, "s", false, "Enable HTTPS")
	flag.StringVar(&commandLineParams.ConfigFile, "c", "", "Config file")
	flag.Parse()

	if params.ConfigFile == "" {
		params.ConfigFile = commandLineParams.ConfigFile
	}
	var fileParams ServerConfig
	if params.ConfigFile != "" {
		fileParams = parseFileParams(params.ConfigFile)
	}

	params.BaseAddress = firstNotZero(params.BaseAddress, commandLineParams.BaseAddress, fileParams.BaseAddress)
	params.ShortURLsAddress = firstNotZero(params.ShortURLsAddress, commandLineParams.ShortURLsAddress, fileParams.ShortURLsAddress)
	params.FileStoragePath = firstNotZero(params.FileStoragePath, commandLineParams.FileStoragePath, fileParams.FileStoragePath)
	params.DatabaseDSN = firstNotZero(params.DatabaseDSN, commandLineParams.DatabaseDSN, fileParams.DatabaseDSN)
	params.EnableHTTPS = firstNotZero(params.EnableHTTPS, commandLineParams.EnableHTTPS, fileParams.EnableHTTPS)

	return &params, nil
}
