package model

import "time"

// Status order yang mungkin terjadi
// pending    → order baru dibuat, belum diproses
// processing → sedang diproses payment
// completed  → payment berhasil
// cancelled  → payment gagal atau dibatalkan
type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted  OrderStatus = "completed"
	StatusCancelled  OrderStatus = "cancelled"
)

// OrderItem adalah satu baris produk di dalam order
type OrderItem struct {
	ID        string
	OrderID   string
	ProductID string
	Quantity  int
	Price     float64
}

// Order adalah keseluruhan pesanan dari satu user
type Order struct {
	ID        string
	UserID    string
	Status    OrderStatus
	Total     float64
	Items     []OrderItem
	CreatedAt time.Time
	UpdatedAt time.Time
}
