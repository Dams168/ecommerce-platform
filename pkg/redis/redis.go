package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client membungkus redis.Client dari library go-redis.
// Semua service akan pakai struct ini — bukan langsung pakai
// library go-redis — supaya kalau nanti mau ganti library,
// cukup ubah di satu tempat ini saja.
type ClientRedis struct {
    rdb *redis.Client // Ganti dengan tipe client Redis yang Anda gunakan, misalnya *redis.Client
}

// config untuk koneksi Redis, bisa diisi dari environment variable atau file konfigurasi
type Config struct {
    Addr string // alamat Redis, misalnya "localhost:6379"
    Password string // password Redis, jika ada
    DB int // nomor database Redis, biasanya 0
}

func NewClientRedis(cfg Config) (*ClientRedis, error) {
    rdb := redis.NewClient(&redis.Options{
        Addr: cfg.Addr,
        Password: cfg.Password,
        DB: cfg.DB,
    })

    // Coba ping untuk memastikan koneksi berhasil
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := rdb.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("failed to connect to %s Redis: %w", cfg.Addr, err)
    }

    return &ClientRedis{rdb: rdb}, nil
}

// operasi dasar Redis, misalnya Set dan Get

// Set menyimpan nilai di Redis dengan key tertentu dan waktu kedaluwarsa
func (c *ClientRedis) Set(ctx context.Context, key, value string, ttl time.Duration) error {
    return c.rdb.Set(ctx, key, value, ttl).Err()
}

// Get mengambil nilai dari Redis berdasarkan key
func (c *ClientRedis) Get(ctx context.Context, key string) (string, error) {
    return c.rdb.Get(ctx, key).Result()
}

// Delete menghapus key dari Redis
func (c *ClientRedis) Delete(ctx context.Context, key string) error {
    return c.rdb.Del(ctx, key).Err()
}

//exists func untuk cek apakah key ada di Redis
func (c *ClientRedis) Exists(ctx context.Context, key string) (bool, error) {
    count, err := c.rdb.Exists(ctx, key).Result()
    return count > 0, err
}

// JWT Blacklist -> fungsi untuk menambahkan token yang sudah logout ke blacklist Redis
// AddToBlacklist menambahkan token ke blacklist dengan waktu kedaluwarsa yang sama dengan token itu sendiri
func (c *ClientRedis) AddToBlacklist(ctx context.Context, tokenID string, ttl time.Duration) error {
    key := "blacklist:" + tokenID
    return c.rdb.Set(ctx, key, "1", ttl).Err()
}

// IsBlacklisted memeriksa apakah token ada di blacklist Redis
func (c *ClientRedis) IsBlacklisted(ctx context.Context, tokenID string) (bool, error) {
    key := "blacklist:" + tokenID
    return c.Exists(ctx, key)
}

// ─── Rate Limiting -> fungsi untuk membatasi jumlah permintaan dari user tertentu dalam jangka waktu tertentu

// Dipakai untuk hitung berapa kali user request dalam satu window waktu.
// Mengembalikan jumlah request saat ini.

func (c *ClientRedis) IncrementWithExpiry(ctx context.Context, key string, ttl time.Duration) (int64, error) {
    pipe := c.rdb.Pipeline()
    incr := pipe.Incr(ctx, key)
    pipe.Expire(ctx, key, ttl)
    _, err := pipe.Exec(ctx)
    if err != nil {
        return 0, err
    }
    return incr.Val(), nil
}

// ───── Pub/sub (Notifikasi Websocket) -> fungsi untuk mengirim notifikasi real-time ke user melalui WebSocket

// Publish mengirim pesan ke channel tertentu
// semua subscriber yang subscribe ke channel ini akan menerima pesan ini
// Dipakai Notification service untuk trigger push ke WebSocket

func (c *ClientRedis) Publish(ctx context.Context, channel, message string) error {
    return c.rdb.Publish(ctx, channel, message).Err()
}

// Subscribe berlangganan ke channel tertentu untuk menerima pesan
func (c *ClientRedis) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
    return c.rdb.Subscribe(ctx, channels...)
}

// Close menutup koneksi Redis
func (c *ClientRedis) Close() error {
    return c.rdb.Close()
}