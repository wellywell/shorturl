package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	BaseAddress      string `env:"SERVER_ADDRESS"`
	ShortURLsAddress string `env:"BASE_URL"`
}

func NewConfig() (*Config, error) {
	var params Config
	err := env.Parse(&params)
	if err != nil {
		return nil, err
	}

	var commandLineParams Config

	flag.StringVar(&commandLineParams.BaseAddress, "a", "localhost:8080", "Base address to listen on")
	flag.StringVar(&commandLineParams.ShortURLsAddress, "b", "http://localhost:8080", "Short URLs base address")
	flag.Parse()

	if params.BaseAddress == "" {
		params.BaseAddress = commandLineParams.BaseAddress
	} else {
		log.Infof("Using ENV param SERVER_ADDRESS: %s", params.BaseAddress)
	}
	if params.ShortURLsAddress == "" {
		params.ShortURLsAddress = commandLineParams.ShortURLsAddress
	} else {
		log.Infof("Using ENV param BASE_URL: %s", params.ShortURLsAddress)
	}

	return &params, nil
}

func (c *Config) GetBaseAddress() string {
	return c.BaseAddress
}

func (c *Config) GetShortURLsAddress() string {
	return c.ShortURLsAddress
}
