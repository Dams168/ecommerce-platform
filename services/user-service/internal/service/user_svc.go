package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Dams168/ecommerce-platform/pkg/jwt"
	"github.com/Dams168/ecommerce-platform/user-service/internal/model"
	"github.com/Dams168/ecommerce-platform/user-service/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
    Register(ctx context.Context, name, email, password string) (*model.User, error)
	Login(ctx context.Context, email, password string) (string, *model.User, error)
	GetUser(ctx context.Context, userID string) (*model.User, error)
}

type UserServiceImpl struct {
    repo          repository.UserRepository
	jwtManager    *jwt.Manager
}

func NewUserService(repo repository.UserRepository, jwtManager *jwt.Manager) UserService {
    return &UserServiceImpl{
        repo:          repo,
        jwtManager: jwtManager,
    }
}
// GetUser implements [UserService].
func (u *UserServiceImpl) GetUser(ctx context.Context, userID string) (*model.User, error) {
	return u.repo.FindByID(ctx, userID)
}

// Login implements [UserService].
func (u *UserServiceImpl) Login(ctx context.Context, email string, password string) (string, *model.User, error) {
	// 1. Cari user berdasarkan email
    user, err := u.repo.FindByEmail(ctx, email)
    if err != nil {
        return "", nil, fmt.Errorf("Email atau password salah")
    }

    // 2. Cek password
    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return "", nil, fmt.Errorf("Email atau password salah")
    }

    // 3. Generate JWT token
    token, err := u.jwtManager.GenerateToken(user.ID, user.Email)
    if err != nil {
        return "", nil, fmt.Errorf("Gagal generate token: %w", err)
    }

    return token, user, nil
}

// Register implements [UserService].
func (u *UserServiceImpl) Register(ctx context.Context, name string, email string, password string) (*model.User, error) {
	// validasi dasar
    if name == "" || email == "" || password == "" {
        return nil, fmt.Errorf("Nama, email, dan password wajib diisi")
    }
    if len(password) < 8 {
        return nil, fmt.Errorf("Password minimal 6 karakter")
    }

    // cek email sudah terdaftar atau belum
    exist, err := u.repo.ExistByEmail(ctx, email)
    if err != nil {
        return nil, err
    }
    if exist {
        return nil, fmt.Errorf("Email sudah terdaftar")
    }

    // hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return nil, fmt.Errorf("Gagal menghash password: %w", err)
    }

    // buat user baru
    user := &model.User{
        ID: uuid.NewString(),
        Name:     name,
        Email:    email,
        Password: string(hashedPassword),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }

    // simpan user ke database
    err = u.repo.CreateUser(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("Gagal mendaftar user: %w", err)
    }

    return user, nil
}

