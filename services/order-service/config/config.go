package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GRPCPort     string
	DatabaseURL  string
	KafkaBrokers []string
}

func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:    getEnv("ORDER_SERVICE_PORT", "50052"),
		DatabaseURL: getEnv("DB_ORDER_URL", ""),
		KafkaBrokers: strings.Split(
			getEnv("KAFKA_BROKERS", "localhost:9092"), ",",
		),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DB_ORDER_URL wajib diisi di .env")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
