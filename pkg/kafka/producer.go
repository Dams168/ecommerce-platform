package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Producer dipakai untuk mengirim ke kafka
// satu producer bisa mengirim ke banyak topik, tapi satu pesan hanya bisa ke satu topik
type Producer struct {
    writer  *kafka.Writer
}

func NewProducer(brokers []string) *Producer {
    writer := &kafka.Writer{
        Addr:     kafka.TCP(brokers...),

        // Hash: pesan dengan key yang sama selalu masuk partisi yang sama.
		// Penting untuk menjaga urutan event dari satu order.
        Balancer: &kafka.Hash{},

        // RequireAll: tunggu semua replica simpan pesan sebelum anggap sukses.
		// Lebih lambat tapi lebih aman — pesan tidak hilang kalau satu broker mati.
        RequiredAcks: kafka.RequireAll,

        // RequireAll: tunggu semua replica simpan pesan sebelum anggap sukses.
		// Lebih lambat tapi lebih aman — pesan tidak hilang kalau satu broker mati.
        Async: false,
        WriteTimeout: 10 * time.Second,
    }
    return &Producer{writer: writer}
}


// Publish mengirim satu event ke topic tertentu.
//
// topic  = nama topic, pakai konstanta dari topics.go
// key    = partition key, biasanya ID entitas (order_id, user_id)
//          pesan dengan key sama → masuk partisi sama → urutan terjaga
// value  = data yang dikirim, akan di-convert ke JSON otomatis
func (p *Producer) Publish(ctx context.Context, topic, key string, value any) error {
	data, err := json.Marshal(value)
    if err != nil {
        return fmt.Errorf("failed to marshal message value: %w", err)
    }

    err = p.writer.WriteMessages(ctx, kafka.Message{
        Topic: topic,
        Key:   []byte(key),
        Value: data,
    })
    if err != nil {
        return fmt.Errorf("failed to write message to Kafka: %w", err)
    }
    return nil
}

// Close menutup koneksi Producer dengan bersih.
// Selalu panggil ini saat service shutdown.
func (p *Producer) Close() error {
    return p.writer.Close()
}