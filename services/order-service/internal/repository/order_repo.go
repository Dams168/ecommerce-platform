package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dams168/ecommerce-platform/order-service/internal/model"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

// Save menyimpan order beserta semua item-nya dalam satu transaksi.
// Kalau salah satu gagal, semuanya dibatalkan — tidak ada data setengah tersimpan.
func (r *OrderRepository) Save(ctx context.Context, order *model.Order) error {
	// mulai transaksi — semua query di bawah masuk satu paket
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("gagal mulai transaksi: %w", err)
	}
	// kalau ada panic atau return error, otomatis rollback
	defer tx.Rollback(ctx)

	// simpan order dulu
	_, err = tx.Exec(ctx, `
		INSERT INTO orders (id, user_id, status, total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`, order.ID, order.UserID, order.Status, order.Total)
	if err != nil {
		return fmt.Errorf("gagal simpan order: %w", err)
	}

	// simpan semua item
	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4, $5)
		`, item.ID, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return fmt.Errorf("gagal simpan item: %w", err)
		}
	}

	// commit — tandai transaksi berhasil
	return tx.Commit(ctx)
}

// FindByID mengambil order beserta semua item-nya
func (r *OrderRepository) FindByID(ctx context.Context, orderID string) (*model.Order, error) {
	// ambil data order
	order := &model.Order{}
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders WHERE id = $1
	`, orderID).Scan(
		&order.ID, &order.UserID, &order.Status,
		&order.Total, &order.CreatedAt, &order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("order tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal ambil order: %w", err)
	}

	// ambil semua item order ini
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, product_id, quantity, price
		FROM order_items WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, fmt.Errorf("gagal ambil items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(&item.ID, &item.OrderID,
			&item.ProductID, &item.Quantity, &item.Price)
		if err != nil {
			return nil, err
		}
		order.Items = append(order.Items, item)
	}

	return order, nil
}

// FindByUserID mengambil semua order milik satu user
func (r *OrderRepository) FindByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Order, int, error) {
	// hitung total dulu
	var total int
	r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID,
	).Scan(&total)

	// ambil data dengan pagination
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, status, total, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("gagal ambil orders: %w", err)
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		rows.Scan(&order.ID, &order.UserID, &order.Status,
			&order.Total, &order.CreatedAt, &order.UpdatedAt)
		orders = append(orders, order)
	}

	return orders, total, nil
}

// UpdateStatus mengubah status order
// Dipanggil oleh payment-worker setelah payment selesai
func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID string, status model.OrderStatus) error {
	_, err := r.db.Exec(ctx, `
		UPDATE orders SET status = $1, updated_at = NOW()
		WHERE id = $2
	`, status, orderID)
	return err
}
