package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Dams168/ecommerce-platform/api-gateway/internal/middleware"
	userv1 "github.com/Dams168/ecommerce-platform/proto/gen/user/v1"
)

type AuthHandler struct {
	userClient userv1.UserServiceClient
	logger     *slog.Logger
}

func NewAuthHandler(userClient userv1.UserServiceClient, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{userClient: userClient, logger: logger}
}

// Register — POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	resp, err := h.userClient.Register(r.Context(), &userv1.RegisterRequest{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.InvalidArgument {
				writeError(w, http.StatusBadRequest, st.Message())
				return
			}
		}
		writeError(w, http.StatusInternalServerError, "gagal register")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user_id": resp.UserId,
		"name":    resp.Name,
		"email":   resp.Email,
	})
}

// Login — POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "request tidak valid")
		return
	}

	resp, err := h.userClient.Login(r.Context(), &userv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.Unauthenticated {
				writeError(w, http.StatusUnauthorized, st.Message())
				return
			}
		}
		writeError(w, http.StatusInternalServerError, "gagal login")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"token":   resp.Token,
		"user_id": resp.UserId,
		"name":    resp.Name,
	})
}

// GetMe — GET /users/me
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// ambil userID dari context — diset oleh Auth middleware
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	resp, err := h.userClient.GetUser(r.Context(), &userv1.GetUserRequest{
		UserId: userID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "user tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": resp.UserId,
		"name":    resp.Name,
		"email":   resp.Email,
	})
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}