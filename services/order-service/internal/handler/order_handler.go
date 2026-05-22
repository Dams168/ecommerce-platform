package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Dams168/ecommerce-platform/order-service/internal/service"
	orderv1 "github.com/Dams168/ecommerce-platform/proto/gen/order/v1"
)

type OrderHandler struct {
	orderv1.UnimplementedOrderServiceServer
	svc *service.OrderService
}

func New(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

func (h *OrderHandler) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id wajib diisi")
	}

	// convert proto items ke service input
	items := make([]service.ItemInput, 0, len(req.Items))
	for _, i := range req.Items {
		items = append(items, service.ItemInput{
			ProductID: i.ProductId,
			Quantity:  int(i.Quantity),
			Price:     i.Price,
		})
	}

	order, err := h.svc.CreateOrder(ctx, service.CreateOrderInput{
		UserID: req.UserId,
		Items:  items,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &orderv1.CreateOrderResponse{
		OrderId: order.ID,
		Status:  string(order.Status),
		Total:   order.Total,
	}, nil
}

func (h *OrderHandler) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	order, err := h.svc.GetOrder(ctx, req.OrderId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, err.Error())
	}

	// convert items ke proto format
	protoItems := make([]*orderv1.OrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		protoItems = append(protoItems, &orderv1.OrderItem{
			ProductId: item.ProductID,
			Quantity:  int32(item.Quantity),
			Price:     item.Price,
		})
	}

	return &orderv1.GetOrderResponse{
		OrderId: order.ID,
		UserId:  order.UserID,
		Status:  string(order.Status),
		Total:   order.Total,
		Items:   protoItems,
	}, nil
}

func (h *OrderHandler) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	orders, total, err := h.svc.ListOrders(
		ctx, req.UserId, int(req.Limit), int(req.Offset),
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	protoOrders := make([]*orderv1.GetOrderResponse, 0, len(orders))
	for _, o := range orders {
		protoOrders = append(protoOrders, &orderv1.GetOrderResponse{
			OrderId: o.ID,
			UserId:  o.UserID,
			Status:  string(o.Status),
			Total:   o.Total,
		})
	}

	return &orderv1.ListOrdersResponse{
		Orders: protoOrders,
		Total:  int32(total),
	}, nil
}
