package ws

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// GetUpgrader mengembalikan upgrader untuk dipakai handler
func GetUpgrader() *websocket.Upgrader {
	return &upgrader
}

// Hub menyimpan semua koneksi WebSocket aktif.
// Key = userID, Value = koneksi WebSocket user tersebut.
type Hub struct {
	clients map[string]*websocket.Conn
	mu      sync.RWMutex
	rdb     *redis.Client
	logger  *slog.Logger
}

func NewHub(rdb *redis.Client, logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[string]*websocket.Conn),
		rdb:     rdb,
		logger:  logger,
	}
}

// Register menyimpan koneksi WebSocket user ke Hub.
func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	// kalau user sudah ada koneksi lama, tutup dulu
	if old, ok := h.clients[userID]; ok {
		old.Close()
	}
	h.clients[userID] = conn
	h.mu.Unlock()
	h.logger.Info("user terhubung via WebSocket", "user_id", userID)
}

// Unregister menghapus koneksi WebSocket user dari Hub.
func (h *Hub) Unregister(userID string) {
	h.mu.Lock()
	delete(h.clients, userID)
	h.mu.Unlock()
	h.logger.Info("user terputus dari WebSocket", "user_id", userID)
}

// SendToUser mengirim pesan ke user tertentu via WebSocket.
func (h *Hub) SendToUser(userID string, msg []byte) {
	h.mu.RLock()
	conn, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		// user tidak sedang online — tidak apa-apa
		// notifikasi sudah tersimpan di database
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		h.logger.Warn("gagal kirim pesan WebSocket",
			"user_id", userID,
			"error", err,
		)
		h.Unregister(userID)
	}
}

// SubscribeRedis mendengarkan Redis Pub/Sub untuk semua channel notif:*
// dan forward pesan ke WebSocket user yang tepat.
// Dipanggil sekali saat startup — berjalan terus di background.
func (h *Hub) SubscribeRedis(ctx context.Context) {
	// subscribe ke pattern notif:* — semua channel notifikasi
	pubsub := h.rdb.PSubscribe(ctx, "notif:*")
	defer pubsub.Close()

	h.logger.Info("Hub mulai subscribe Redis Pub/Sub")

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// extract userID dari nama channel
			// format channel: "notif:{userID}"
			userID := msg.Channel[len("notif:"):]

			h.logger.Info("menerima notifikasi dari Redis",
				"user_id", userID,
				"payload", msg.Payload,
			)

			// forward ke WebSocket user
			h.SendToUser(userID, []byte(msg.Payload))
		}
	}
}