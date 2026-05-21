package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Dams168/ecommerce-platform/pkg/jwt"
	"github.com/Dams168/ecommerce-platform/pkg/logger"
	userv1 "github.com/Dams168/ecommerce-platform/proto/gen/user/v1"
	"github.com/Dams168/ecommerce-platform/user-service/config"
	"github.com/Dams168/ecommerce-platform/user-service/internal/handler"
	"github.com/Dams168/ecommerce-platform/user-service/internal/repository"
	"github.com/Dams168/ecommerce-platform/user-service/internal/service"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	_ "github.com/golang-migrate/migrate/v4/database/postgres" // ← harus ada
	_ "github.com/golang-migrate/migrate/v4/source/file"
)


func main() {
    // inisialisasi logger
    log := logger.New("user-service")

    // load konfigurasi
    cfg, err := config.Load()
    if err != nil {
        log.Error("Gagal memuat konfigurasi", "error", err)
        os.Exit(1)
    }

    // inisialisasi database
    log.Info("Menghubungkan ke database...")
    db, err := pgxpool.New(context.Background(), cfg.DatabaseUrl)
    if err != nil {
        log.Error("Gagal menghubungkan ke database", "error", err)
        os.Exit(1)
    }
    defer db.Close()

    //verifikasi koneksi database
    if err := db.Ping(context.Background()); err != nil {
        log.Error("Gagal memverifikasi koneksi database", "error", err)
        os.Exit(1)
    }
    log.Info("Berhasil terhubung ke database")

    //migrasi database
    log.Info("menjalankan migrasi database...")

    migrationPath := "file:///home/dam/projects/ecommerce-platform/services/user-service/migrations"
    log.Info("migration path", "path", migrationPath)

    m, err := migrate.New(migrationPath, cfg.DatabaseUrl)
    if err != nil {
        log.Error("gagal inisialisasi migration", "error", err)
        os.Exit(1)
    }
    if m == nil {
        log.Error("migrate.New() return nil tanpa error")
        os.Exit(1)
    }

    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        log.Error("gagal jalankan migration", "error", err)
        os.Exit(1)
    }
    log.Info("migration selesai")

    // inisialisasi repository, service, handler
    jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTexpiration)
    userRepo := repository.NewUserRepository(db)
    userSvc := service.NewUserService(userRepo, jwtManager)
    userHandler := handler.NewUserHandler(userSvc)

    // jalankan gRPC server
    grpcServer := grpc.NewServer()
    userv1.RegisterUserServiceServer(grpcServer, userHandler)

    reflection.Register(grpcServer)

    addr := fmt.Sprintf(":%s", cfg.GRPCPort)
    listener, err := net.Listen("tcp", addr)
    if err != nil {
        log.Error("Gagal membuat listener gRPC", "error", err)
        os.Exit(1)
    }

    // graceful shutdown — tangkap sinyal Ctrl+C atau kill
	// agar request yang sedang diproses bisa selesai dulu
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// jalankan server di goroutine terpisah
	go func() {
		log.Info("user-service berjalan", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("server error", "error", err)
		}
	}()

	// tunggu sinyal shutdown
	<-ctx.Done()
	log.Info("mematikan server...")
	grpcServer.GracefulStop()
	log.Info("server berhenti dengan bersih")

}