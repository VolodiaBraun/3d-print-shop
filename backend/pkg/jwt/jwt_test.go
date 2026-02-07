package jwt

import (
	"testing"
	"time"
)

func newTestManager() *Manager {
	return NewManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
}

func TestGenerateTokenPair(t *testing.T) {
	m := newTestManager()

	pair, err := m.GenerateTokenPair(42, "customer")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("access token should not be empty")
	}
	if pair.RefreshToken == "" {
		t.Error("refresh token should not be empty")
	}
	if pair.AccessToken == pair.RefreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

func TestValidateAccessToken_Valid(t *testing.T) {
	m := newTestManager()

	pair, err := m.GenerateTokenPair(42, "admin")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := m.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if claims.UserID != 42 {
		t.Errorf("expected userID 42, got %d", claims.UserID)
	}
	if claims.Role != "admin" {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	// Create a manager with very short expiry
	m := NewManager("test-secret-key", -1*time.Second, 7*24*time.Hour)

	pair, err := m.GenerateTokenPair(1, "customer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = m.ValidateAccessToken(pair.AccessToken)
	if err != ErrTokenExpired {
		t.Errorf("expected ErrTokenExpired, got %v", err)
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	m := newTestManager()

	_, err := m.ValidateAccessToken("not-a-real-token")
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	m1 := NewManager("secret-one", 15*time.Minute, 7*24*time.Hour)
	m2 := NewManager("secret-two", 15*time.Minute, 7*24*time.Hour)

	pair, err := m1.GenerateTokenPair(1, "customer")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = m2.ValidateAccessToken(pair.AccessToken)
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAccessExpiry(t *testing.T) {
	m := newTestManager()
	if m.AccessExpiry() != 15*time.Minute {
		t.Errorf("expected 15m, got %v", m.AccessExpiry())
	}
}

func TestRefreshExpiry(t *testing.T) {
	m := newTestManager()
	if m.RefreshExpiry() != 7*24*time.Hour {
		t.Errorf("expected 168h, got %v", m.RefreshExpiry())
	}
}
