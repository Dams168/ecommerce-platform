package redis

import (
	"context"
	"testing"
	"time"
)

func TestClientRedis(t *testing.T) {
    ctx := context.Background()

    client, err := NewClientRedis(Config{
        Addr: "localhost:6379",
        Password: "",
        DB: 0,
    })
    if err != nil {
        t.Fatalf("Failed to create Redis client: %v", err)
    }
    defer client.Close()
    t.Log("Connected to Redis successfully")

    // Test Set and Get

    err = client.Set(ctx, "test:key", "Hello, Redis!", 1*time.Minute)
    if err != nil {
        t.Fatalf("Failed to set key: %v", err)
    }

    value, err := client.Get(ctx, "test:key")
    if err != nil {
        t.Fatalf("Failed to get key: %v", err)
    }
    if value != "Hello, Redis!" {
        t.Errorf("Expected 'Hello, Redis!', got '%s'", value)
    }
    t.Logf("Set and Get operations successful : %s", value)

    // Test Blacklist token

    err = client.AddToBlacklist(ctx, "token-abc123", 1*time.Minute)
    if err != nil {
        t.Fatalf("Failed to add token to blacklist: %v", err)
    }
    
    blocked, err := client.IsBlacklisted(ctx, "token-abc123")
    if err != nil || !blocked {
        t.Fatalf("Failed to check blacklist: %v", err)
    }
    t.Log("Blacklist operations successful, token is blacklisted")

    // Rate Limiting

    count, err := client.IncrementWithExpiry(ctx, "rate-limit:user123", 1*time.Minute)
    if err != nil {
        t.Fatalf("Failed to increment rate limit: %v", err)
    }
    t.Logf("Rate limit count for user123: %d", count)

    // Delete key
    err = client.Delete(ctx, "test:key")
    if err != nil {
        t.Fatalf("Failed to delete key: %v", err)
    }

    exists, err := client.Exists(ctx, "test:key")
    if err != nil {
        t.Fatalf("Failed to check key existence: %v", err)
    }
    t.Logf("Key existence after deletion: %v", exists)
}

