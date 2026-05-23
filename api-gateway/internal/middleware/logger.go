package middleware

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// responseWriter wrapper untuk capture status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack mengimplementasikan http.Hijacker
// dibutuhkan agar WebSocket bisa upgrade koneksi
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: 200}

			next.ServeHTTP(rw, r)

			log.Info("request",
				"method",     r.Method,
				"path",       r.URL.Path,
				"status",     rw.status,
				"latency_ms", time.Since(start).Milliseconds(),
				"request_id", r.Context().Value(RequestIDKey),
			)
		})
	}
}