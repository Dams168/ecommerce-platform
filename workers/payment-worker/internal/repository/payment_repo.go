package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Payment struct {
	ID          string
	OrderID     string
	UserID      string
	Amount      float64
	Status      string
	ProcessedAt *time.Time
	CreatedAt   time.Time
}

type PaymentRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// IsProcessed mengecek apakah order ini sudah pernah diproses.
// Ini adalah cek idempotency — mencegah double payment.
// Kafka bisa kirim pesan yang sama lebih dari sekali (at-least-once),
// jadi kita harus cek dulu sebelum proses.
func (r *PaymentRepository) IsProcessed(ctx context.Context, orderID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM payments WHERE order_id = $1)`,
		orderID,
	).Scan(&exists)
	return exists, err
}

// Save menyimpan record pembayaran baru.
func (r *PaymentRepository) Save(ctx context.Context, p *Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments (id, order_id, user_id, amount, status, processed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`, p.ID, p.OrderID, p.UserID, p.Amount, p.Status, p.ProcessedAt)
	if err != nil {
		return fmt.Errorf("gagal simpan payment: %w", err)
	}
	return nil
}