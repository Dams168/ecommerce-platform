package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DatabaseURL  string
	KafkaBrokers []string
	RedisAddr    string
	RedisPassword string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: getEnv("DB_NOTIFICATION_URL", ""),
		KafkaBrokers: strings.Split(
			getEnv("KAFKA_BROKERS", "localhost:9092"), ",",
		),
		RedisAddr:    getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DB_NOTIFICATION_URL wajib diisi")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}