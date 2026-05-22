package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Dams168/ecommerce-platform/order-service/config"
	"github.com/Dams168/ecommerce-platform/order-service/internal/event"
	"github.com/Dams168/ecommerce-platform/order-service/internal/handler"
	"github.com/Dams168/ecommerce-platform/order-service/internal/repository"
	"github.com/Dams168/ecommerce-platform/order-service/internal/service"
	"github.com/Dams168/ecommerce-platform/pkg/logger"
	orderv1 "github.com/Dams168/ecommerce-platform/proto/gen/order/v1"
)

func main() {
	log := logger.New("order-service")

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
		"file:///home/dam/projects/ecommerce-platform/services/order-service/migrations",
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

	// rangkai semua layer
	publisher := event.NewPublisher(cfg.KafkaBrokers)
	defer publisher.Close()

	orderRepo := repository.New(db)
	orderSvc := service.New(orderRepo, publisher)
	orderHandler := handler.New(orderSvc)

	// gRPC server
	grpcServer := grpc.NewServer()
	orderv1.RegisterOrderServiceServer(grpcServer, orderHandler)
	reflection.Register(grpcServer)

	addr := fmt.Sprintf(":%s", cfg.GRPCPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("gagal buka port", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("order-service berjalan", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	log.Info("mematikan server...")
	grpcServer.GracefulStop()
	log.Info("server berhenti dengan bersih")
	_ = log // suppress unused warning
}
