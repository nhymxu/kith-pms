package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_Roundtrip(t *testing.T) {
	plain := "correct-horse-battery-staple"
	hash, err := HashPassword(plain)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Errorf("unexpected hash prefix: %q", hash[:10])
	}

	ok, err := VerifyPassword(hash, plain)
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if !ok {
		t.Error("expected password to verify correctly")
	}
}

func TestHashPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}

	ok, err := VerifyPassword(hash, "wrong")
	if err != nil {
		t.Fatalf("VerifyPassword: %v", err)
	}
	if ok {
		t.Error("expected wrong password to fail verification")
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	plain := "same-password"
	h1, _ := HashPassword(plain)
	h2, _ := HashPassword(plain)
	if h1 == h2 {
		t.Error("two hashes of the same password should differ (different salts)")
	}
}

func TestVerifyPassword_InvalidHashFormat(t *testing.T) {
	_, err := VerifyPassword("not-a-valid-hash", "anything")
	if err == nil {
		t.Error("expected error for invalid hash format")
	}
}
