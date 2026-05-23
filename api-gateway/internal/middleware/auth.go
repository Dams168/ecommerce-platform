package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Dams168/ecommerce-platform/pkg/jwt"
)

// Auth memvalidasi JWT token di setiap request.
// Kalau valid, userID disimpan di context untuk dipakai handler.
// Kalau tidak valid, langsung return 401.
func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// ambil token dari header Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "token tidak ada")
				return
			}

			// format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "format token salah")
				return
			}

			// verifikasi token
			claims, err := jwtManager.VerifyToken(parts[1])
			if err != nil {
				writeError(w, http.StatusUnauthorized, "token tidak valid")
				return
			}

			// simpan userID di context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// writeError menulis response error dalam format JSON
func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// writeJSON menulis response sukses dalam format JSON
func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}