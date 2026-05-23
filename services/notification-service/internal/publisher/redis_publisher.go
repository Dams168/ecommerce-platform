package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NotificationMessage adalah format pesan yang dikirim
// ke Redis Pub/Sub. API Gateway akan terima ini
// dan forward ke WebSocket client.
type NotificationMessage struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	OrderID string `json:"order_id,omitempty"`
}

type RedisPublisher struct {
	client *redis.Client
}

func New(addr, password string) *RedisPublisher {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
	})
	return &RedisPublisher{client: client}
}

// Publish mengirim notifikasi ke channel Redis milik user.
//
// Channel format: notif:{user_id}
// API Gateway subscribe ke channel ini saat user connect WebSocket.
// Ketika ada pesan masuk, Gateway langsung push ke WebSocket user.
func (p *RedisPublisher) Publish(ctx context.Context, userID string, msg NotificationMessage) error {
	// convert ke JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("gagal marshal notifikasi: %w", err)
	}

	// publish ke channel khusus user ini
	channel := fmt.Sprintf("notif:%s", userID)
	if err := p.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("gagal publish ke Redis channel %s: %w", channel, err)
	}

	return nil
}

func (p *RedisPublisher) Close() error {
	return p.client.Close()
}