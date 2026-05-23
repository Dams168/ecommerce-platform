package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Dams168/ecommerce-platform/api-gateway/internal/config"
	"github.com/Dams168/ecommerce-platform/api-gateway/internal/handler"
	"github.com/Dams168/ecommerce-platform/api-gateway/internal/middleware"
	"github.com/Dams168/ecommerce-platform/api-gateway/internal/ws"
	"github.com/Dams168/ecommerce-platform/pkg/jwt"
	"github.com/Dams168/ecommerce-platform/pkg/logger"
	orderv1 "github.com/Dams168/ecommerce-platform/proto/gen/order/v1"
	userv1 "github.com/Dams168/ecommerce-platform/proto/gen/user/v1"
)

func main() {
	log := logger.New("api-gateway")

	cfg, err := config.Load()
	if err != nil {
		log.Error("gagal load config", "error", err)
		os.Exit(1)
	}

	// ── gRPC connections ──────────────────────────────────────
	// konek ke user-service
	log.Info("menghubungkan ke user-service...", "addr", cfg.UserServiceAddr)
	userConn, err := grpc.NewClient(cfg.UserServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("gagal konek user-service", "error", err)
		os.Exit(1)
	}
	defer userConn.Close()

	// konek ke order-service
	log.Info("menghubungkan ke order-service...", "addr", cfg.OrderServiceAddr)
	orderConn, err := grpc.NewClient(cfg.OrderServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("gagal konek order-service", "error", err)
		os.Exit(1)
	}
	defer orderConn.Close()

	// buat gRPC client untuk tiap service
	userClient  := userv1.NewUserServiceClient(userConn)
	orderClient := orderv1.NewOrderServiceClient(orderConn)

	// ── Redis ─────────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Error("gagal konek Redis", "error", err)
		os.Exit(1)
	}
	log.Info("Redis terhubung")

	// ── JWT ───────────────────────────────────────────────────
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTExpiration)

	// ── WebSocket Hub ─────────────────────────────────────────
	hub := ws.NewHub(rdb, log)

	// jalankan Redis subscriber di background
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go hub.SubscribeRedis(ctx)

	// ── Handlers ──────────────────────────────────────────────
	authHandler  := handler.NewAuthHandler(userClient, log)
	orderHandler := handler.NewOrderHandler(orderClient, log)
	wsHandler    := handler.NewWSHandler(hub, jwtManager, log)

	// ── Router ────────────────────────────────────────────────
	r := chi.NewRouter()

	// middleware global — dieksekusi untuk semua request
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger(log))
	r.Use(chiMiddleware.Recoverer) // recover dari panic

	// ── Routes publik — tidak perlu JWT ───────────────────────
	r.Post("/auth/register", authHandler.Register)
	r.Post("/auth/login",    authHandler.Login)

	// WebSocket — auth dilakukan di dalam handler via query param
	r.Get("/ws", wsHandler.HandleWS)

	// ── Routes private — wajib JWT ────────────────────────────
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(jwtManager))

		r.Get("/users/me", authHandler.GetMe)

		r.Post("/orders",     orderHandler.CreateOrder)
		r.Get("/orders/{id}", orderHandler.GetOrder)
	})

	// ── HTTP Server ───────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Info("api-gateway berjalan",
			"port", cfg.Port,
			"routes", []string{
				"POST /auth/register",
				"POST /auth/login",
				"GET  /users/me",
				"POST /orders",
				"GET  /orders/{id}",
				"GET  /ws",
			},
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
		}
	}()

	// tunggu sinyal shutdown
	<-ctx.Done()
	log.Info("mematikan server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("gagal shutdown server", "error", err)
	}

	log.Info("api-gateway berhenti dengan bersih")
}