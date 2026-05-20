package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
)

// HandlerFunc adalah tipe fungsi yang dipanggil untuk setiap pesan yang diterima.
// setiap service mendefinisikan handler sendiri sesuai kebutuhan
// kalau handler error, pesan akan retry
type HandlerFunc func(ctx context.Context, msg kafka.Message) error

type Consumer struct {
    reader *kafka.Reader
    handler HandlerFunc
    logger *slog.Logger
}

// NewConsumer membuat Consumer baru.
//
// brokers = alamat Kafka
// topic   = topic yang mau dibaca
// groupID = nama consumer group (pakai konstanta dari topics.go)
// handler = fungsi yang dipanggil untuk setiap pesan masuk
// logger  = untuk mencatat aktivitas consumer

func NewConsumer(brokers []string, topic, groupID string, handler HandlerFunc, logger *slog.Logger) *Consumer {
    reader := kafka.NewReader(kafka.ReaderConfig{
        Brokers: brokers,
        Topic:   topic,
        GroupID: groupID,

        // MinBytes: tunggu sampai minimal 1 byte tersedia
		// MaxBytes: maksimal ambil 10MB per fetch
        MinBytes: 1,
        MaxBytes: 10e6, // 10MB


        // CommitInterval 0 = manual commit
		// Kita commit sendiri setelah pesan berhasil diproses
		// Kalau auto-commit, pesan bisa hilang kalau service crash sebelum proses selesai
        CommitInterval: 0,

        // Kalau consumer baru pertama kali join,
		// mulai dari pesan terbaru (bukan dari awal)
        StartOffset: kafka.LastOffset,

        // Kalau consumer tidak kirim heartbeat > 30 detik,
		// Kafka anggap consumer mati dan rebalance partisi
        SessionTimeout: 30 * time.Second,
    })
    return &Consumer{
        reader: reader,
        handler: handler,
        logger: logger,
    }
}

// Run memulai loop membaca pesan. Berhenti saat ctx di-cancel.
// Biasanya dipanggil di goroutine tersendiri.
//
// Alurnya:
// 1. Ambil pesan dari Kafka (FetchMessage)
// 2. Panggil handler untuk proses pesan
// 3. Kalau sukses → commit offset (beritahu Kafka pesan sudah diproses)
// 4. Kalau gagal → retry maksimal 3x, lalu skip kalau tetap gagal

func (c *Consumer) Run(ctx context.Context) {
    defer c.reader.Close()

    c.logger.Info("Kafka consumer started",
        "topic", c.reader.Config().Topic,
        "group", c.reader.Config().GroupID,
    )
    
    for {
        // FetchMessage block (menunggu) sampai ada pesan baru
		// Langsung return kalau ctx di-cancel (saat shutdown)
        msg, err := c.reader.FetchMessage(ctx)
        if err != nil {
            // ctx di-cancel = shutdown normal, bukan error
            if errors.Is(err, context.Canceled) {
                c.logger.Info("Kafka consumer stopped")
                return 
            }
            c.logger.Error("Failed to fetch message", "error", err)
            time.Sleep(time.Second)
            continue
        }

        c.logger.Info("Message received", 
            "topic", msg.Topic, 
            "partition", msg.Partition, 
            "offset", msg.Offset,
        )

        // Proses pesan dengan retry otomatis
        if err := c.processWithRetry(ctx, msg, 3); err != nil {
			// setelah 3x retry masih gagal, skip pesan ini
			// agar tidak block pesan berikutnya
			c.logger.Error("pesan gagal diproses setelah 3x retry, dilewati",
				"offset", msg.Offset,
				"error", err,
			)
		}
        // commit offset = beritahu Kafka bahwa pesan ini sudah diproses
		// Kafka tidak akan kirim pesan ini lagi ke consumer group kita
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Error("gagal commit offset", "error", err)
		}
    }
}

func (c *Consumer) processWithRetry(ctx context.Context, msg kafka.Message, maxRetry int) error {
	for attempt := range maxRetry {
		err := c.handler(ctx, msg)
		if err == nil {
			return nil // sukses, langsung keluar
		}

		// jeda exponential: 1s, 2s, 4s
		wait := time.Duration(1<<attempt) * time.Second
		c.logger.Warn("gagal memproses pesan, akan retry",
			"offset", msg.Offset,
			"attempt", attempt+1,
			"error", err,
		)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}

	return fmt.Errorf("gagal setelah %d percobaan", maxRetry)
}

// Unmarshal adalah helper untuk decode JSON dari pesan Kafka.
// Dipakai di dalam handler: kafka.Unmarshal(msg.Value, &myStruct)
func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}