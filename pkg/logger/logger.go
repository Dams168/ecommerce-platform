package logger

import (
	"log/slog"
	"os"
)

// Fungsi New membuat logger baru dengan nama Service yang diberikan
// misalnya user-service, order-service, dll.
// Tujuannya untuk memudahkan identifikasi log dari service mana.
func New(serviceName string) *slog.Logger{
    handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    })
    logger := slog.New(handler).With("service", serviceName)
    return logger
}