package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: getEnv("DB_PAYMENT_URL", ""),
		KafkaBrokers: strings.Split(
			getEnv("KAFKA_BROKERS", "localhost:9092"), ",",
		),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DB_PAYMENT_URL wajib diisi di .env")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}