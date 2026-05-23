package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"github.com/Dams168/ecommerce-platform/payment-worker/internal/repository"
	kafkapkg "github.com/Dams168/ecommerce-platform/pkg/kafka"
)

// OrderCreatedEvent adalah struktur pesan yang diterima dari Kafka.
// Harus sama persis dengan yang di-publish order-service.
type OrderCreatedEvent struct {
	OrderID   string      `json:"order_id"`
	UserID    string      `json:"user_id"`
	Total     float64     `json:"total"`
	Items     []EventItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
}

type EventItem struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// PaymentDoneEvent adalah event yang kita publish setelah payment selesai.
// Notification-service akan consume event ini.
type PaymentDoneEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

type PaymentHandler struct {
	repo     *repository.PaymentRepository
	producer *kafkapkg.Producer
	logger   *slog.Logger
}

func New(
	repo *repository.PaymentRepository,
	producer *kafkapkg.Producer,
	logger *slog.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		repo:     repo,
		producer: producer,
		logger:   logger,
	}
}

// Handle dipanggil untuk setiap pesan order.created dari Kafka.
//
// Alur:
// 1. Parse pesan JSON
// 2. Cek idempotency — sudah diproses?
// 3. Simulasi proses payment
// 4. Simpan ke database
// 5. Publish event payment.done
func (h *PaymentHandler) Handle(ctx context.Context, msg kafka.Message) error {
	// 1. parse pesan
	var event OrderCreatedEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		// pesan corrupt — skip, jangan retry
		// return nil supaya tidak di-retry terus
		h.logger.Error("pesan tidak bisa di-parse, dilewati",
			"error", err,
			"offset", msg.Offset,
		)
		return nil
	}

	h.logger.Info("memproses payment",
		"order_id", event.OrderID,
		"user_id", event.UserID,
		"total", event.Total,
	)

	// 2. cek idempotency
	// kalau sudah diproses sebelumnya, skip tanpa error
	// ini penting karena Kafka bisa kirim pesan yang sama 2x
	processed, err := h.repo.IsProcessed(ctx, event.OrderID)
	if err != nil {
		return fmt.Errorf("gagal cek idempotency: %w", err)
	}
	if processed {
		h.logger.Info("order sudah diproses sebelumnya, dilewati",
			"order_id", event.OrderID,
		)
		return nil
	}

	// 3. simulasi proses payment
	// di production: panggil payment gateway (Midtrans, Stripe, dll)
	// untuk project ini kita simulasikan selalu berhasil
	now := time.Now()
	status := "completed"

	h.logger.Info("payment berhasil",
		"order_id", event.OrderID,
		"amount", event.Total,
	)

	// 4. simpan ke database
	err = h.repo.Save(ctx, &repository.Payment{
		ID:          uuid.NewString(),
		OrderID:     event.OrderID,
		UserID:      event.UserID,
		Amount:      event.Total,
		Status:      status,
		ProcessedAt: &now,
	})
	if err != nil {
		return fmt.Errorf("gagal simpan payment: %w", err)
	}

	// 5. publish event payment.done
	// notification-service akan consume ini untuk kirim notifikasi ke user
	doneEvent := PaymentDoneEvent{
		OrderID:     event.OrderID,
		UserID:      event.UserID,
		Amount:      event.Total,
		Status:      status,
		ProcessedAt: now,
	}

	if err := h.producer.Publish(
		ctx,
		kafkapkg.TopicPaymentDone,
		event.OrderID,
		doneEvent,
	); err != nil {
		// log warning tapi tidak gagalkan — payment sudah tersimpan
		h.logger.Warn("gagal publish payment.done",
			"order_id", event.OrderID,
			"error", err,
		)
	}

	h.logger.Info("payment selesai diproses",
		"order_id", event.OrderID,
		"status", status,
	)

	return nil
}