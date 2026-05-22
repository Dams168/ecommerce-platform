package event

import (
	"context"
	"time"

	kafkapkg "github.com/Dams168/ecommerce-platform/pkg/kafka"
)

// OrderCreatedEvent adalah struktur pesan yang dikirim ke Kafka.
// Payment Worker akan menerima dan memproses event ini.
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

// Publisher bertanggung jawab publish event order ke Kafka
type Publisher struct {
	producer *kafkapkg.Producer
}

func NewPublisher(brokers []string) *Publisher {
	return &Publisher{
		producer: kafkapkg.NewProducer(brokers),
	}
}

// PublishOrderCreated mengirim event order.created ke Kafka.
// Dipanggil setelah order berhasil disimpan ke database.
// key = order_id supaya semua event dari order yang sama
// masuk ke partisi yang sama — urutan terjaga.
func (p *Publisher) PublishOrderCreated(ctx context.Context, event OrderCreatedEvent) error {
	return p.producer.Publish(
		ctx,
		kafkapkg.TopicOrderCreated,
		event.OrderID,
		event,
	)
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}
