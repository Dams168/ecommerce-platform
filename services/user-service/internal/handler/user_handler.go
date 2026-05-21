package handler

import (
	"context"

	userv1 "github.com/Dams168/ecommerce-platform/proto/gen/user/v1"
	"github.com/Dams168/ecommerce-platform/user-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandlerImpl struct {
    userv1.UnimplementedUserServiceServer
    svc service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandlerImpl {
    return &UserHandlerImpl{svc: svc}
}

func (h *UserHandlerImpl) Register(ctx context.Context, req *userv1.RegisterRequest) (*userv1.RegisterResponse, error) {
    user, err := h.svc.Register(ctx, req.Name, req.Email, req.Password)
    if err != nil {
        // gRPC punya status code sendiri — lebih spesifik dari HTTP
		// InvalidArgument = input dari client salah
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
    }
    return &userv1.RegisterResponse{
        UserId: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    }, nil
}

func (h *UserHandlerImpl) Login(ctx context.Context, req *userv1.LoginRequest) (*userv1.LoginResponse, error) {
    token, user, err := h.svc.Login(ctx, req.Email, req.Password)
    if err != nil {
        // Unauthenticated = token salah
        return nil, status.Errorf(codes.Unauthenticated, err.Error())
    }
    return &userv1.LoginResponse{
        UserId: user.ID,
        Name:   user.Name,
        Token:  token,
    }, nil
}

func (h *UserHandlerImpl) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
    user, err := h.svc.GetUser(ctx, req.UserId)
    if err != nil {
        // NotFound = data tidak ditemukan
        return nil, status.Errorf(codes.NotFound, err.Error())
    }
    return &userv1.GetUserResponse{
        UserId: user.ID,
        Name:   user.Name,
        Email:  user.Email,
    }, nil
}

