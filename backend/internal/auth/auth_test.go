package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "s3cret-password" {
		t.Fatal("password was stored in plaintext")
	}
	if !CheckPassword(hash, "s3cret-password") {
		t.Error("CheckPassword rejected the correct password")
	}
	if CheckPassword(hash, "wrong-password") {
		t.Error("CheckPassword accepted an incorrect password")
	}
}

func TestJWTRoundTrip(t *testing.T) {
	m := NewManager("test-secret", time.Hour)
	id := uuid.New()

	token, err := m.Generate(id, "admin")
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	claims, err := m.Parse(token)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if claims.UserID != id {
		t.Errorf("UserID = %v, want %v", claims.UserID, id)
	}
	if claims.Role != "admin" {
		t.Errorf("Role = %q, want admin", claims.Role)
	}
}

func TestJWTRejectsTamperedAndExpired(t *testing.T) {
	m := NewManager("test-secret", time.Hour)

	if _, err := m.Parse("not-a-real-token"); err == nil {
		t.Error("Parse accepted a garbage token")
	}

	// A token signed with a different secret must be rejected.
	other := NewManager("different-secret", time.Hour)
	token, _ := other.Generate(uuid.New(), "user")
	if _, err := m.Parse(token); err == nil {
		t.Error("Parse accepted a token signed with the wrong secret")
	}

	// An already-expired token must be rejected.
	expired := NewManager("test-secret", -time.Minute)
	token, _ = expired.Generate(uuid.New(), "user")
	if _, err := m.Parse(token); err == nil {
		t.Error("Parse accepted an expired token")
	}
}
