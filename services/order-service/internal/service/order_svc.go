package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Dams168/ecommerce-platform/order-service/internal/event"
	"github.com/Dams168/ecommerce-platform/order-service/internal/model"
	"github.com/Dams168/ecommerce-platform/order-service/internal/repository"
)

type OrderService struct {
	repo      *repository.OrderRepository
	publisher *event.Publisher
}

func New(repo *repository.OrderRepository, publisher *event.Publisher) *OrderService {
	return &OrderService{repo: repo, publisher: publisher}
}

type CreateOrderInput struct {
	UserID string
	Items  []ItemInput
}

type ItemInput struct {
	ProductID string
	Quantity  int
	Price     float64
}

// CreateOrder menyimpan order ke DB lalu publish event ke Kafka.
//
// Urutan ini penting:
// 1. Simpan ke DB dulu — kalau Kafka down, order tetap tersimpan
// 2. Baru publish ke Kafka — payment worker akan proses async
//
// Kalau dibalik (publish dulu baru simpan), dan DB gagal,
// payment worker sudah proses order yang tidak ada di DB — kacau.
func (s *OrderService) CreateOrder(ctx context.Context, input CreateOrderInput) (*model.Order, error) {
	if len(input.Items) == 0 {
		return nil, fmt.Errorf("order harus punya minimal 1 item")
	}

	// hitung total dan buat item
	var total float64
	items := make([]model.OrderItem, 0, len(input.Items))
	eventItems := make([]event.EventItem, 0, len(input.Items))

	for _, i := range input.Items {
		if i.Quantity <= 0 {
			return nil, fmt.Errorf("quantity harus lebih dari 0")
		}
		if i.Price <= 0 {
			return nil, fmt.Errorf("harga harus lebih dari 0")
		}

		itemID := uuid.NewString()
		total += float64(i.Quantity) * i.Price

		items = append(items, model.OrderItem{
			ID:        itemID,
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			Price:     i.Price,
		})

		eventItems = append(eventItems, event.EventItem{
			ProductID: i.ProductID,
			Quantity:  i.Quantity,
			Price:     i.Price,
		})
	}

	order := &model.Order{
		ID:     uuid.NewString(),
		UserID: input.UserID,
		Status: model.StatusPending,
		Total:  total,
		Items:  items,
	}

	// 1. simpan ke database
	if err := s.repo.Save(ctx, order); err != nil {
		return nil, fmt.Errorf("gagal simpan order: %w", err)
	}

	// 2. publish event ke Kafka
	// kalau gagal, kita log warning tapi tidak gagalkan order
	// ada mekanisme retry/reconciliation yang bisa dibangun nanti
	err := s.publisher.PublishOrderCreated(ctx, event.OrderCreatedEvent{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Total:     order.Total,
		Items:     eventItems,
		CreatedAt: time.Now(),
	})
	if err != nil {
		// LOG warning — order tetap berhasil dibuat
		fmt.Printf("WARNING: gagal publish event untuk order %s: %v\n", order.ID, err)
	}

	return order, nil
}

func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	return s.repo.FindByID(ctx, orderID)
}

func (s *OrderService) ListOrders(ctx context.Context, userID string, limit, offset int) ([]*model.Order, int, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.FindByUserID(ctx, userID, limit, offset)
}