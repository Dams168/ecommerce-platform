package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Dams168/ecommerce-platform/api-gateway/internal/middleware"
	orderv1 "github.com/Dams168/ecommerce-platform/proto/gen/order/v1"
)

type OrderHandler struct {
	orderClient orderv1.OrderServiceClient
	logger      *slog.Logger
}

func NewOrderHandler(orderClient orderv1.OrderServiceClient, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{orderClient: orderClient, logger: logger}
}

// CreateOrder — POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		Items []struct {
			ProductID string  `json:"product_id"`
			Quantity  int32   `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	// convert ke proto format
	items := make([]*orderv1.OrderItem, 0, len(req.Items))
	for _, i := range req.Items {
		items = append(items, &orderv1.OrderItem{
			ProductId: i.ProductID,
			Quantity:  i.Quantity,
			Price:     i.Price,
		})
	}

	resp, err := h.orderClient.CreateOrder(r.Context(), &orderv1.CreateOrderRequest{
		UserId: userID,
		Items:  items,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal buat order")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"order_id": resp.OrderId,
		"status":   resp.Status,
		"total":    resp.Total,
	})
}

// GetOrder — GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")

	resp, err := h.orderClient.GetOrder(r.Context(), &orderv1.GetOrderRequest{
		OrderId: orderID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "order tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"order_id": resp.OrderId,
		"user_id":  resp.UserId,
		"status":   resp.Status,
		"total":    resp.Total,
	})
}