package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Notification struct {
	ID     string
	UserID string
	Type   string
	Title  string
	Body   string
}

type NotificationRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Save menyimpan notifikasi ke database.
// Tujuannya: user bisa lihat riwayat notifikasi
// walau mereka sedang offline saat notifikasi dikirim.
func (r *NotificationRepository) Save(ctx context.Context, n *Notification) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO notifications (id, user_id, type, title, body, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, uuid.NewString(), n.UserID, n.Type, n.Title, n.Body)
	if err != nil {
		return fmt.Errorf("gagal simpan notifikasi: %w", err)
	}
	return nil
}