package jwt

import (
	"testing"
	"time"
)

func TestGenerateandVerify(t *testing.T){
    manager := NewManager("inirahasiapakebanget", 1*time.Hour)
    token, err := manager.GenerateToken("user123", "user@mail.com")
    if err != nil {
        t.Fatalf("Error generating token: %v", err)
    }

    t.Logf("Generated Token: %s", token[:20]) // Log sebagian token untuk keamanan

    claims, err := manager.VerifyToken(token)
    if err != nil {
        t.Fatalf("Error verifying token: %v", err)
    }
    
    if claims.UserID != "user123" || claims.Email != "user@mail.com" {
        t.Errorf("Expected claims not found")
    }

    t.Logf("Token valid, userID: %s, email: %s", claims.UserID, claims.Email)

    expiredManager := NewManager("inirahasiapakebanget", -1*time.Hour)
    expiredToken, _ := expiredManager.GenerateToken("user123", "user@mail.com")
    _, err = expiredManager.VerifyToken(expiredToken)
    if err == nil {
        t.Fatal("Token Expired ditolak")
    }

    t.Log("token expired berhasil ditolak")
}