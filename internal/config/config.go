package config

import (
	"flag"
	"os"
)

// Config holds the service configuration.
type Config struct {
	Addr string
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Addr: ":8080",
	}
}

// Load reads configuration from defaults < env < flags.
func Load() *Config {
	c := Default()

	// Env
	if addr := os.Getenv("URLKIT_ADDR"); addr != "" {
		c.Addr = addr
	}

	// Flags
	flag.StringVar(&c.Addr, "addr", c.Addr, "listen address")
	flag.Parse()

	return c
}
