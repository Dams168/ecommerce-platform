package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
)

func TestProducerConsumer(t *testing.T) {
	brokers := []string{"localhost:9092"}
	topic := "test.topic"
	group := "test-group"

	type TestEvent struct {
		Message string `json:"message"`
		SentAt  string `json:"sent_at"`
	}

	received := make(chan TestEvent, 1)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	handler := func(ctx context.Context, msg kafka.Message) error {
		var e TestEvent
		if err := Unmarshal(msg.Value, &e); err != nil {
			return err
		}
		received <- e
		return nil
	}

	// ── 1. Start consumer DULU sebelum publish ───────────────
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	consumer := NewConsumer(brokers, topic, group, handler, logger)
	go consumer.Run(ctx)

	// tunggu consumer siap — join ke Kafka dan dapat partisi
	time.Sleep(3 * time.Second)
	t.Log("consumer siap")

	// ── 2. Baru publish pesan ─────────────────────────────────
	producer := NewProducer(brokers)
	defer producer.Close()

	event := TestEvent{
		Message: "halo dari producer",
		SentAt:  time.Now().Format(time.RFC3339),
	}

	if err := producer.Publish(ctx, topic, "test-key", event); err != nil {
		t.Fatalf("gagal publish: %v", err)
	}
	t.Log("pesan berhasil dikirim ke Kafka")

	// ── 3. Tunggu pesan diterima consumer ────────────────────
	select {
	case e := <-received:
		data, _ := json.MarshalIndent(e, "", "  ")
		t.Logf("pesan diterima:\n%s", data)
	case <-ctx.Done():
		t.Fatal("timeout — pesan tidak diterima dalam 20 detik")
	}
}