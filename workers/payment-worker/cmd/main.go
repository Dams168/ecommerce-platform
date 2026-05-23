package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dams168/ecommerce-platform/payment-worker/config"
	"github.com/Dams168/ecommerce-platform/payment-worker/internal/handler"
	"github.com/Dams168/ecommerce-platform/payment-worker/internal/repository"
	kafkapkg "github.com/Dams168/ecommerce-platform/pkg/kafka"
	"github.com/Dams168/ecommerce-platform/pkg/logger"
)

func main() {
	log := logger.New("payment-worker")

	cfg, err := config.Load()
	if err != nil {
		log.Error("gagal load config", "error", err)
		os.Exit(1)
	}

	// database
	log.Info("menghubungkan ke database...")
	db, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Error("gagal konek database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Error("database tidak bisa di-ping", "error", err)
		os.Exit(1)
	}
	log.Info("database terhubung")

	// migration
	log.Info("menjalankan migration...")
	m, err := migrate.New(
		"file:///home/dam/projects/ecommerce-platform/workers/payment-worker/migrations",
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Error("gagal inisialisasi migration", "error", err)
		os.Exit(1)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Error("gagal jalankan migration", "error", err)
		os.Exit(1)
	}
	log.Info("migration selesai")

	// producer untuk publish payment.done
	producer := kafkapkg.NewProducer(cfg.KafkaBrokers)
	defer producer.Close()

	// handler
	paymentRepo := repository.New(db)
	paymentHandler := handler.New(paymentRepo, producer, log)

	// consumer — mulai dengarkan Kafka
	consumer := kafkapkg.NewConsumer(
		cfg.KafkaBrokers,
		kafkapkg.TopicOrderCreated,
		kafkapkg.GroupPaymentworker,
		paymentHandler.Handle,
		log,
	)

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("payment-worker siap, mendengarkan Kafka...",
		"topic", kafkapkg.TopicOrderCreated,
		"group", kafkapkg.GroupPaymentworker,
	)

	// run blocking — berhenti saat ctx di-cancel
	consumer.Run(ctx)

	log.Info("payment-worker berhenti dengan bersih")
}