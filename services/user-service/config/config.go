package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
    GRPCPort string

    DatabaseUrl string // format: postgres://username:password@host:port/dbname
    JWTSecret   string
    JWTexpiration time.Duration
}

// Load Membaca semua kofigurasi dari .env
func Load() (*Config, error) {
    cfg := &Config{
        GRPCPort: getEnv("GRPC_PORT", "50051"),

        DatabaseUrl: getEnv("DB_USER_URL", ""),
        JWTSecret:   getEnv("JWT_SECRET", ""),
        JWTexpiration: 24 * time.Hour,
    }

    // validasi — field wajib tidak boleh kosong
	if cfg.DatabaseUrl == "" {
		return nil, fmt.Errorf("DB_USER_URL wajib diisi di .env")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET wajib diisi di .env")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if  val := os.Getenv(key); val != "" {
        return val
    }
    return defaultValue
}