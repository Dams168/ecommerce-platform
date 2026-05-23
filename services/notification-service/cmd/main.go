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

	"github.com/Dams168/ecommerce-platform/notification-service/config"
	notifconsumer "github.com/Dams168/ecommerce-platform/notification-service/internal/consumer"
	"github.com/Dams168/ecommerce-platform/notification-service/internal/publisher"
	"github.com/Dams168/ecommerce-platform/notification-service/internal/repository"
	kafkapkg "github.com/Dams168/ecommerce-platform/pkg/kafka"
	"github.com/Dams168/ecommerce-platform/pkg/logger"
)

func main() {
	log := logger.New("notification-service")

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
		"file:///home/dam/projects/ecommerce-platform/services/notification-service/migrations",
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

	// redis publisher
	redisPublisher := publisher.New(cfg.RedisAddr, cfg.RedisPassword)
	defer redisPublisher.Close()

	// rangkai handler
	notifRepo := repository.New(db)
	notifHandler := notifconsumer.New(notifRepo, redisPublisher, log)

	// kafka consumer — dengarkan payment.done
	consumer := kafkapkg.NewConsumer(
		cfg.KafkaBrokers,
		kafkapkg.TopicPaymentDone,
		kafkapkg.GroupNotificationSvc,
		notifHandler.Handle,
		log,
	)

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("notification-service siap, mendengarkan Kafka...",
		"topic", kafkapkg.TopicPaymentDone,
		"group", kafkapkg.GroupNotificationSvc,
	)

	consumer.Run(ctx)
	log.Info("notification-service berhenti dengan bersih")
}