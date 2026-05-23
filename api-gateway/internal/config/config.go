package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port                    string
	JWTSecret               string
	JWTExpiration           time.Duration
	UserServiceAddr         string
	OrderServiceAddr        string
	RedisAddr               string
	RedisPassword           string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:             getEnv("GATEWAY_PORT", "8080"),
		JWTSecret:        getEnv("JWT_SECRET", ""),
		JWTExpiration:    24 * time.Hour,
		UserServiceAddr:  getEnv("USER_SERVICE_ADDR", "localhost:50051"),
		OrderServiceAddr: getEnv("ORDER_SERVICE_ADDR", "localhost:50052"),
		RedisAddr:        getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    getEnv("REDIS_PASSWORD", ""),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}