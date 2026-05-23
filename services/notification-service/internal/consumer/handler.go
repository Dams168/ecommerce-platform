package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"

	"github.com/Dams168/ecommerce-platform/notification-service/internal/publisher"
	"github.com/Dams168/ecommerce-platform/notification-service/internal/repository"
)

// PaymentDoneEvent adalah struktur pesan dari payment-worker.
// Harus sama dengan yang di-publish payment-worker.
type PaymentDoneEvent struct {
	OrderID     string    `json:"order_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

type NotificationHandler struct {
	repo      *repository.NotificationRepository
	publisher *publisher.RedisPublisher
	logger    *slog.Logger
}

func New(
	repo *repository.NotificationRepository,
	pub *publisher.RedisPublisher,
	logger *slog.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		repo:      repo,
		publisher: pub,
		logger:    logger,
	}
}

// Handle dipanggil untuk setiap pesan payment.done dari Kafka.
//
// Alur:
// 1. Parse event payment.done
// 2. Tentukan judul dan isi notifikasi berdasarkan status
// 3. Simpan notifikasi ke database (untuk riwayat)
// 4. Publish ke Redis Pub/Sub (untuk real-time push)
func (h *NotificationHandler) Handle(ctx context.Context, msg kafka.Message) error {
	// 1. parse event
	var event PaymentDoneEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		h.logger.Error("pesan tidak bisa di-parse, dilewati",
			"error", err,
			"offset", msg.Offset,
		)
		return nil // skip pesan corrupt
	}

	h.logger.Info("menerima event payment",
		"order_id", event.OrderID,
		"user_id", event.UserID,
		"status", event.Status,
	)

	// 2. tentukan konten notifikasi berdasarkan status payment
	var title, body string
	switch event.Status {
	case "completed":
		title = "Pembayaran Berhasil"
		body = fmt.Sprintf(
			"Pembayaran order #%s sebesar Rp %.0f berhasil diproses.",
			event.OrderID[:8], event.Amount,
		)
	case "failed":
		title = "Pembayaran Gagal"
		body = fmt.Sprintf(
			"Pembayaran order #%s gagal. Silakan coba lagi.",
			event.OrderID[:8],
		)
	default:
		title = "Status Order Diperbarui"
		body = fmt.Sprintf("Order #%s: %s", event.OrderID[:8], event.Status)
	}

	notif := &repository.Notification{
		UserID: event.UserID,
		Type:   "payment." + event.Status,
		Title:  title,
		Body:   body,
	}

	// 3. simpan ke database
	if err := h.repo.Save(ctx, notif); err != nil {
		h.logger.Error("gagal simpan notifikasi", "error", err)
		// lanjutkan — tetap coba kirim real-time
	}

	// 4. publish ke Redis Pub/Sub untuk real-time push
	redisMsg := publisher.NotificationMessage{
		Type:    notif.Type,
		Title:   notif.Title,
		Body:    notif.Body,
		OrderID: event.OrderID,
	}

	if err := h.publisher.Publish(ctx, event.UserID, redisMsg); err != nil {
		h.logger.Warn("gagal publish ke Redis",
			"user_id", event.UserID,
			"error", err,
		)
	}

	h.logger.Info("notifikasi berhasil dikirim",
		"user_id", event.UserID,
		"title", title,
	)

	return nil
}