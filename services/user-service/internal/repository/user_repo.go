package repository

import (
	"context"
	"fmt"

	"github.com/Dams168/ecommerce-platform/user-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	ExistByEmail(ctx context.Context, email string) (bool, error)
}

type UserRepositoryImpl struct {
    db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
    return &UserRepositoryImpl{db: db}
}
// CreateUser implements [UserRepository].
func (u *UserRepositoryImpl) CreateUser(ctx context.Context, user *model.User) error {
	query := `
        INSERT INTO users (id, name, email, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, NOW(), NOW())`
	_, err := u.db.Exec(ctx, query, user.ID, user.Name, user.Email, user.Password)
	if err != nil {
        return fmt.Errorf("Gagal menyimpan user : %w", err)
    }
        
    return nil   
}

// ExistByEmail implements [UserRepository].
func (u *UserRepositoryImpl) ExistByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

    var exists bool
    err := u.db.QueryRow(ctx, query, email).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("Gagal mengecek email: %w", err)
    }
    return exists, nil

}

// FindByEmail implements [UserRepository].
func (u *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = $1`

    user := model.User{}
    err := u.db.QueryRow(ctx, query, email).Scan(
        &user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("Gagal mencari user berdasarkan email: %w", err)
    }
    return &user, nil
}

// FindByID implements [UserRepository].
func (u *UserRepositoryImpl) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = $1`

    user := model.User{}
    err := u.db.QueryRow(ctx, query, id).Scan(
        &user.ID, &user.Name, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("Gagal mencari user berdasarkan ID: %w", err)
    }
    return &user, nil
}

