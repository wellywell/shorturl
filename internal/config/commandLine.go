package config

import (
	"flag"
)

type CommandLineParams struct {
	BaseAddress      string
	ShortURLsAddress string
}

func NewCommandLineParams() *CommandLineParams {
	params := &CommandLineParams{}

	flag.StringVar(&params.BaseAddress, "a", "localhost:8080", "Base address to listen on")
	flag.StringVar(&params.ShortURLsAddress, "b", "localhost:8080", "Short URLs base address")
	flag.Parse()

	return params
}

func (c *CommandLineParams) GetBaseAddress() string {
	return c.BaseAddress
}

func (c *CommandLineParams) GetShortURLsAddress() string {
	return c.ShortURLsAddress
}
