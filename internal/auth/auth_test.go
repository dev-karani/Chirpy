package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID :=uuid.New()
	secret := "my-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT failed: %v", err)
	}

	gotID, err := ValidateJWT(token, secret)
	if err !=nil {
		t.Fatalf("ValidateJWT failed,%v", err)
	}
	
	if gotID != userID {
		t.Errorf("expected %v,%v", userID, gotID)
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