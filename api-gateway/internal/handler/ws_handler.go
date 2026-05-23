package handler

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/Dams168/ecommerce-platform/api-gateway/internal/ws"
	"github.com/Dams168/ecommerce-platform/pkg/jwt"
)

type WSHandler struct {
	hub        *ws.Hub
	jwtManager *jwt.Manager
	logger     *slog.Logger
}

func NewWSHandler(hub *ws.Hub, jwtManager *jwt.Manager, logger *slog.Logger) *WSHandler {
	return &WSHandler{hub: hub, jwtManager: jwtManager, logger: logger}
}

// HandleWS — GET /ws
// Client konek dengan token di query param:
// ws://localhost:8080/ws?token=eyJ...
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	// ambil token dari query param
	// kenapa query param? karena WebSocket tidak support
	// custom header saat handshake di browser
	tokenStr := r.URL.Query().Get("token")
	if tokenStr == "" {
		// coba dari header Authorization
		auth := r.Header.Get("Authorization")
		if auth != "" {
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) == 2 {
				tokenStr = parts[1]
			}
		}
	}

	if tokenStr == "" {
		http.Error(w, "token tidak ada", http.StatusUnauthorized)
		return
	}

	// validasi JWT
	claims, err := h.jwtManager.VerifyToken(tokenStr)
	if err != nil {
		http.Error(w, "token tidak valid", http.StatusUnauthorized)
		return
	}

	// upgrade HTTP → WebSocket
	conn, err := ws.GetUpgrader().Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("gagal upgrade WebSocket", "error", err)
		return
	}
	defer conn.Close()

	userID := claims.UserID

	// daftarkan koneksi ke Hub
	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID)

	// kirim pesan selamat datang
	conn.WriteJSON(map[string]string{
		"type":    "connected",
		"message": "WebSocket terhubung, siap menerima notifikasi",
	})

	// loop — baca pesan dari client (untuk heartbeat)
	// kalau client disconnect, loop ini akan error dan keluar
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseNormalClosure,
			) {
				h.logger.Warn("WebSocket disconnect tidak normal",
					"user_id", userID,
					"error", err,
				)
			}
			break
		}
	}
}
