package config

import "os"

const (
	DefaultAddr    = ":4566"
	DefaultDataDir = "./data"
)

type Config struct {
	Addr    string
	DataDir string
}

func FromEnvironment() Config {
	return Config{
		Addr:    valueOrDefault("EMULITH_ADDR", DefaultAddr),
		DataDir: valueOrDefault("EMULITH_DATA_DIR", DefaultDataDir),
	}
}

func valueOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
