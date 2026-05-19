package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims struct yang digunakan untuk menyimpan informasi yang akan dimasukkan ke dalam token JWT.
type Claims struct {
    UserID string `json:"user_id"`
    Email string `json:"email"`

    jwt.RegisteredClaims
}

// Manager struct yang akan digunakan untuk mengelola pembuatan dan validasi token JWT.
type Manager struct {
    secret []byte
    expiration time.Duration
}

// NewManager membuat instance baru dari Manager dengan secret dan expiration yang diberikan.
func NewManager(secret string, expiration time.Duration) *Manager {
    return &Manager{
        secret: []byte(secret),
        expiration: expiration,
    }
}

// GenerateToken membuat token JWT baru dengan informasi userID dan email yang diberikan.
func (m *Manager) GenerateToken(userID, email string) (string, error) {
    now := time.Now()
    
    claims := &Claims{
        UserID: userID,
        Email: email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(now.Add(m.expiration)),
            IssuedAt: jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    
    signed, err := token.SignedString(m.secret)
    if err != nil {
        return "", err
    }

    return signed, nil
}

// ValidateToken memvalidasi token JWT yang diberikan dan mengembalikan claims jika token valid.
func (m *Manager) VerifyToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(
        tokenString,
        &Claims{},
        func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("Algoritma Token Tidak Valid ")
            }
            return m.secret, nil
        },
    )
    if err != nil {
        return nil, err
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("Token Tidak Valid")
    }

    return claims, nil
}
