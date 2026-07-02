package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer abc123")

	got, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := "abc123"
	if got != want {
		t.Errorf("got %q, want %q,", got, want)
	}
}
func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "my-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT failed,%v", err)
	}

	if gotID != userID {
		t.Errorf("got %v, want %v", userID, gotID)
	}
}

func TestExpiredJWT(t *testing.T) {
	userID := uuid.New()
	secret := "my-secret"

	token, _ := MakeJWT(userID, secret, -time.Hour)

	_, err := ValidateJWT(token, secret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}
}

func TestWrongSecretJWT(t *testing.T) {
	userID := uuid.New()
	token, _ := MakeJWT(userID, "correct-secret", time.Hour)

	_, err := ValidateJWT(token, "wrong-secret")
	if err == nil {
		t.Error("expected error for wrong secret, got nil")
	}
}
